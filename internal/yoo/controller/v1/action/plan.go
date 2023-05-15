package action

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/sync/errgroup"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/known"
	"phos.cc/yoo/internal/pkg/log"
	veldt "phos.cc/yoo/internal/pkg/validator"
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/socket_client"
	"phos.cc/yoo/internal/yoo/storage"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

const CACHE_ROOT = "/opt/web/ci/.yoo"

type TaskErr struct {
	Message string
	TaskID  int32
}

func (t TaskErr) Error() string {
	return t.Message + " ,Task ID: " + strconv.Itoa(int(t.TaskID))
}

func (ctrl *ActionController) ExecPlan(c *gin.Context) {

	var r *v1.ExecPlanRequest

	if err := c.ShouldBindJSON(&r); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			core.WriteResponse(c, errno.ErrInvalidParameter.SetMessage(veldt.Translate(errs)), nil)
		} else {
			core.WriteResponse(c, errno.ErrBind, nil)
		}
		return
	}

	go func(c context.Context, b biz.Biz, pid int32, gid int32, onlyFailed bool, isMicro bool) {
		execPlan(c, b, pid, gid, onlyFailed, isMicro)
	}(c, ctrl.b, r.PlanID, r.GroupID, r.OnlyFailed, r.IsMicro)

	c.JSON(200, nil)
}

func (ctrl *ActionController) Download(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	// 返回给前端
	fileName := fmt.Sprintf("%s/project-%d/bundles.zip", CACHE_ROOT, pid)
	file, err := os.Open(fileName)
	if err != nil {
		core.WriteResponse(c, errno.ErrBadRequest, nil)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		core.WriteResponse(c, errno.ErrBadRequest, nil)
		return
	}
	fileSize := fileInfo.Size()
	c.Writer.Header().Set("Content-Disposition", "attachment; fileName="+"bundles.zip")
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	http.ServeContent(c.Writer, c.Request, fileName, time.Now(), file)

}

func execPlan(c context.Context, b biz.Biz, pid int32, gid int32, onlyFailed bool, isMicro bool) {
	// 开始时间
	start := time.Now()

	total := int64(math.MaxInt)
	// 成功数量
	var succeed uint32
	// 失败数量
	var failed uint32

	log.C(c).Infow("start exec plan", "pid", pid, "gid", gid)
	email := c.Value(known.XEmailKey).(string)

	// 检查是否有正在执行的计划
	p, err := b.Plans().Get(c, pid)
	if err != nil {
		log.Errorw("get plan failed", "error", err)
		return
	}

	if p.Status == 2 {
		log.Errorw("plan is executing")
		socket_client.WriteJSON(email, gin.H{
			"event": "plan",
			"type":  "error",
			"msg":   "plan is executing, you must wait it finished",
		})
		return
	} else {
		// 更新计划状态
		if err := b.Plans().Update(c, &v1.UpdatePlanRequest{
			Status: 2,
		}, pid); err != nil {
			log.Errorw("update plan status failed", "error", err)
			return
		}
	}

	bundles, www, repos, err := prepareToExec(pid)
	if err != nil {
		log.Errorw("prepare failed, plan can't be executed")
		return
	}

	r := &v1.ListTaskRequest{
		Page:     1,
		PageSize: 5, // 打包速度基本和磁盘性能有关系，故此只开启 5 个协程
		PlanID:   int(pid),
	}

	var nginxsb strings.Builder
	var sqlsb strings.Builder
	nginxsb.WriteString(`
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;
    absolute_redirect off;
    add_header Access-Control-Allow-Origin *;

    #access_log  /var/log/nginx/host.access.log  main;

	location /api/ {
		proxy_pass http://192.168.31.72:8080/;
	}

    location /assets/ {
    	proxy_pass http://192.168.31.72:84/;
    }

	location /api/yoo/ {
    	proxy_pass http://yoo-resource:8080/;
    }

	location /ws/ {
    	proxy_pass http://192.168.31.72:8080/;
    	proxy_http_version 1.1;
    	proxy_set_header Upgrade $http_upgrade;
    	proxy_set_header Connection "Upgrade";
    }
`)
	sqlsb.WriteString(`
# 创建数据库
CREATE DATABASE IF NOT EXISTS yoo CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

# 使用创建的数据库
USE yoo;

CREATE TABLE IF NOT EXISTS ` + "`resources`" + `
(
    ` + "`id`" + `               int PRIMARY KEY AUTO_INCREMENT,
    ` + "`name`" + `             varchar(255) UNIQUE NOT NULL COMMENT '项目名',
    ` + "`description`" + `      varchar(255)        NOT NULL COMMENT '中文说明，一般为项目 description',
    ` + "`badge`" + `            varchar(255)        NOT NULL COMMENT '应用图标',
	` + "`fake`" + `             tinyint(1) NOT NULL DEFAULT 0 COMMENT '应用为无实体应用, 0 不是, 1 是',
	` + "`url`" + `              varchar(255) COMMENT '非实体应用的链接',
    ` + "`created_at`" + `       timestamp           NOT NULL,
    ` + "`updated_at`" + `       timestamp           NOT NULL
);

CREATE TABLE ` + "`menus`" + ` (
  ` + "`id`" + ` int PRIMARY KEY AUTO_INCREMENT,
  ` + "`name`" + ` varchar(255) NOT NULL COMMENT '菜单名称',
  ` + "`icon`" + ` varchar(255) NOT NULL DEFAULT "default.png" COMMENT '图标',
  ` + "`letter`" + ` varchar(255) NOT NULL COMMENT '中文首字母',
  ` + "`menu_type`" + ` tinyint NOT NULL DEFAULT 1 COMMENT '1 目录, 2 页面, 3 外链',
  ` + "`resource_id`" + ` int COMMENT '资源 id',
  ` + "`href`" + ` varchar(255) UNIQUE COMMENT '路径或者外链',
  ` + "`parent_id`" + ` int COMMENT '父级菜单 id, 若为根目录, 则为空',
  ` + "`number`" + ` int NOT NULL COMMENT '菜单顺序编号',
  ` + "`categories`" + ` json COMMENT '分类 id 列表',
  ` + "`created_at`" + ` timestamp NOT NULL,
  ` + "`updated_at`" + ` timestamp NOT NULL
);

ALTER TABLE ` + "`menus`" + `
    ADD FOREIGN KEY (` + "`resource_id`" + `) REFERENCES ` + "`resources`" + ` (` + "`id`" + `);

ALTER TABLE ` + "`menus`" + `
    ADD FOREIGN KEY (` + "`parent_id`" + `) REFERENCES ` + "`menus`" + ` (` + "`id`" + `);

CREATE TABLE ` + "`category`" + ` (
  ` + "`id`" + ` int PRIMARY KEY AUTO_INCREMENT,
  ` + "`name`" + ` varchar(255) NOT NULL COMMENT '名称',
  ` + "`parent_id`" + ` int COMMENT '父级 id',
  ` + "`created_at`" + ` timestamp NOT NULL,
  ` + "`updated_at`" + ` timestamp NOT NULL
);

ALTER TABLE ` + "`category`" + ` ADD FOREIGN KEY (` + "`parent_id`" + `) REFERENCES ` + "`category`" + ` (` + "`id`" + `);
`)

	for int64((r.Page-1)*r.PageSize) < total {
		var list []*v1.ListTaskResponse
		var err error
		list, total, err = b.Tasks().List(c, r)
		if err != nil {
			log.Errorw("list task", "err", err)
			socket_client.WriteJSON(email, gin.H{
				"event": "plan",
				"type":  "error",
				"msg":   err.Error(),
			})
			return
		}
		r.Page += 1
		g, ctx := errgroup.WithContext(context.Background())
		for _, task := range list {
			func(task *v1.ListTaskResponse) {
				g.Go(func() error {
					select {
					case <-ctx.Done():
						// 将任务状态更新成失败
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 3}, task.ID)
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"type":  "error",
							"data": gin.H{
								"id":     task.ID,
								"status": 3,
							},
							"msg": "ok",
						})
						return ctx.Err()
					default:
						// 将任务状态更新成 执行中
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 2}, task.ID)
						// 通过 websocket 发送消息
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"type":  "info",
							"data": gin.H{
								"id":     task.ID,
								"status": 2,
							},
							"msg": "ok",
						})
						location, sql, sha1, err := runBuildFlow(task, b, c, repos, www, onlyFailed)
						if err != nil {
							b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 3}, task.ID)
							socket_client.WriteJSON(email, gin.H{
								"event": "task",
								"type":  "error",
								"data": gin.H{
									"id":     task.ID,
									"status": 3,
								},
								"msg": "ok",
							})
							atomic.AddUint32(&failed, 1)
							return err
						}

						nginxsb.WriteString(location)

						sqlsb.WriteString(sql)

						// 将任务状态更新成成功
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 4, Sha1: sha1}, task.ID)
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"type":  "success",
							"data": gin.H{
								"id":     task.ID,
								"status": 4,
							},
							"msg": "ok",
						})
						atomic.AddUint32(&succeed, 1)
						return nil
					}
				})
			}(task)
		}
		// 等待任务执行
		_ = g.Wait()
	}

	nginxsb.WriteString(`
	error_page   500 502 503 504  /50x.html;

    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
`)

	// sqlsb.WriteString("\nINSERT INTO menus (`name`, `menu_type`, `href`, `number`, `created_at`, `updated_at`) values ('前端资源', 1, '/resource', 1, NOW(), NOW());\nINSERT INTO menus (`name`, `menu_type`, `href`, `number`, `resource_id`, `parent_id`, `created_at`, `updated_at`) values ('资源管理', 2, '/manage', 1, (select id from resources where resources.name = 'yoo-resource'),(select last_insert_id()), NOW(), NOW())")

	// 在 bundles 目录下生成 default.conf
	file, err := os.Create(fmt.Sprintf("%s/default.conf", bundles))
	if err != nil {
		log.Errorw("create default.conf error", "err", err)
		return
	}
	defer file.Close()
	if _, err := file.WriteString(nginxsb.String()); err != nil {
		log.Errorw("write default.conf error", "err", err)
		return
	}

	// 在 bundles 目录下生成 sql
	file, err = os.Create(fmt.Sprintf("%s/init.sql", bundles))
	if err != nil {
		log.Errorw("create sql error", "err", err)
		return
	}
	defer file.Close()
	if _, err := file.WriteString(sqlsb.String()); err != nil {
		log.Errorw("write sql error", "err", err)
		return
	}

	// 在 bundles 目录下生成 docker-compose.yml
	dockersb := `
version: "3.9"
services:
  yoo-mysql:
    image: "mysql:8.0"
    restart: unless-stopped
    ports:
      - "3307:3306"
    environment:
      LANG: C.UTF-8
      MYSQL_ROOT_PASSWORD: root
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./data:/var/lib/mysql
    command: ['--default-authentication-plugin=mysql_native_password', '--character-set-server=utf8mb4', '--collation-server=utf8mb4_general_ci']

  yoo-nginx:
    image: "nginx:latest"
    restart: unless-stopped
    ports:
      - "8989:80"
    volumes:
      - ./default.conf:/etc/nginx/conf.d/default.conf
      - ./www/:/usr/share/nginx/www/
    depends_on:
      - yoo-mysql

  yoo-resource:
    image: "phostann/yoo-resource:0.0.5"
    restart: unless-stopped
    volumes:
      - ./configs/:/app/configs/
      - ./assets/:/opt/assets
    depends_on:
      - yoo-mysql
`
	file, err = os.Create(fmt.Sprintf("%s/docker-compose.yml", bundles))
	if err != nil {
		log.Errorw("create docker-compose.yml error", "err", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(dockersb); err != nil {
		log.Errorw("write docker-compose.yml error", "err", err)
		return
	}

	// 创建打包产物压缩文件
	zipFile := fmt.Sprintf("%s/project-%d/bundles.zip", CACHE_ROOT, pid)
	if err := compressDir(bundles, zipFile); err != nil {
		log.Errorw("Failed to zip the files")
		return
	}

	end := time.Now()

	elapsed := end.Sub(start)

	timeString := formatDuration(elapsed)

	log.Infow("exec plan finished", "total", total, "succeed", succeed, "failed", failed, "time", timeString)

	if failed == 0 {
		b.Plans().Update(c, &v1.UpdatePlanRequest{Status: 4}, pid)
	} else {

		b.Plans().Update(c, &v1.UpdatePlanRequest{Status: 3}, pid)
	}

	var tp string

	if failed == 0 {
		tp = "success"
	} else {
		tp = "info"
	}

	socket_client.WriteJSON(email, gin.H{
		"event": "plan",
		"type":  tp,
		"data": gin.H{
			"elapsed": timeString,
			"total":   total,
			"succeed": succeed,
			"failed":  failed,
		},
		"msg": "exec plan finished",
	})
}

func prepareToExec(pid int32) (string, string, string, error) {
	// 检查本次打包缓存目录 home
	home := fmt.Sprintf("%s/project-%d", CACHE_ROOT, pid)
	// 检查 缓存目录是否存在, 不存在则创建
	log.Debugw("create home dir", "home", CACHE_ROOT)
	if err := mkDir(home); err != nil {
		log.Errorw("create home dir", "err", err)
		return "", "", "", err
	}

	bundles := fmt.Sprintf("%s/bundles", home)

	// 检查 home 下的 bundles 目录是否存在, 不存在则创建
	log.Debugw("create bundles dir", "bundles", bundles)
	if err := mkDir(bundles); err != nil {
		log.Errorw("create bundles dir", "err", err)
		return "", "", "", err
	}

	www := fmt.Sprintf("%s/www", bundles)

	// 检查 bundles 目录下的 www 目录是否存在，不存在则创建
	log.Infow("create www dir", "www", www)
	if err := mkDir(www); err != nil {
		log.Errorw("create www dir", "err", err)
		return "", "", "", err
	}

	repos := fmt.Sprintf("%s/repos", home)

	// 检查 home 目录下的 repos 目录是否存在，不存在则创建
	log.Infow("create repos dir", "repos", repos)
	if err := mkDir(repos); err != nil {
		log.Errorw("create repos dir", "err", err)
		return "", "", "", err
	}

	return bundles, www, repos, nil
}

func runBuildFlow(task *v1.ListTaskResponse, b biz.Biz, c context.Context, repos string, www string, onlyFailed bool) (string, string, string, error) {
	var output string
	// 是否需要重新构建
	var needBuild bool
	var sha1 string

	if task.Status == 3 {
		needBuild = true
		log.Infow("在上次打包中任务失败，需要重新打包", "project_id", task.ProjectID)
	}

	// 查询 project
	p, err := b.Projects().Get(c, task.ProjectID)
	if err != nil {
		log.Errorw("failed to get project", "project_id", task.ProjectID, "err", err)
		return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	location := fmt.Sprintf(`
	location /%s/{
	    alias /usr/share/nginx/www/%s/;
	    index index.html index.htm;
	    try_files $uri /%s/index.html;
    }
`, p.Name, p.Name, p.Name)

	sql := fmt.Sprintf("\nINSERT INTO `resources` (`name`, `description`, `badge`, `fake`, `created_at`, `updated_at`) VALUES ('%s', '%s', '%s', 0, NOW(), NOW());", p.Name, p.Description, "default.png")

	dir := fmt.Sprintf("%s/%s", repos, p.Name)

	// 比较 git 提交记录，判断是否需要执行构建
	// 优先使用 ssh url
	var repo string
	if p.SSHURL != "" {
		repo = p.SSHURL
	} else {
		repo = p.HTTPURL
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		log.Errorw("get home dir", "project", p.Name, "err", err)
		return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", fmt.Sprintf("%s/.ssh/id_ed25519", homeDir), "")
	if err != nil {
		log.Errorw("get public keys failed", "project", p.Name, "err", err)
		return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	refs, err := rem.List(&git.ListOptions{
		Auth: publicKeys,
	})

	if err != nil {
		log.Errorw("list remote failed", "project", p.Name, "err", err)
		return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	for _, ref := range refs {
		if ref.Name().String() == "refs/heads/master" {
			sha1 = ref.Hash().String()
			break
		}
	}

	// 如果 task 中记录的 sha1 和 git 上项目对应的 sha1 不匹配，则需要执行打包
	if sha1 != task.Sha1 {
		needBuild = true
	}

	// 如果有新代码，或者上次构建中失败，则需要重新构建打包项目
	if needBuild || task.Status == 3 {
		// 删除项目目录
		removeDir(dir)

		// 重新克隆项目
		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			Auth: publicKeys,
			URL:  repo,
		})
		if err != nil {
			log.Errorw("project clone failed", "project", p.Name, "err", err)
			return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
		}

		// 安装依赖
		if output, err = installDeps(dir); err != nil {

			log.Errorw("install dependencies failed", "project", p.Name, "err", err, "output", string(output))
			return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
		}

		// 执行打包
		if output, err = buildProject(dir, p.BuildCmd); err != nil {
			log.Errorw("build project failed", "project", p.Name, "err", err, "output", string(output))
			return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
		}

		// 拷贝打包出的文件至 www 下
		src := fmt.Sprintf("%s/%s", dir, p.Dist)
		dst := fmt.Sprintf("%s/%s", www, p.Name)

		if err := copyDir(src, dst); err != nil {
			log.Errorw("copy dir failed", "project", p.Name, "err", err)
			return "", "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
		}

		// 更新 task 的 sha1 值
	}

	return location, sql, sha1, nil

	// // 如果项目目录不存在，则创建
	// if !isDirExist(dir) {
	// 	needBuild = true
	// 	log.Infow("project dir is not exist, need rebuild", "project", p.Name, "dir", dir)
	// 	log.Debugw("create project dir", "project", p.Name, "dir", dir)
	// 	if err := mkDir(dir); err != nil {
	// 		log.Errorw("create project dir", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}
	// }

	// // 检查当前目录是否为 git 目录
	// gitRepo, err := git.PlainOpen(dir)

	// // 如果不是 git 仓库
	// if err != nil {
	// 	needBuild = true
	// 	log.Infow("repo is not exist, need rebuild", "project", p.Name, "dir", dir)
	// 	// 检查目录是否为空
	// 	empty, err := isDirEmpty(dir)
	// 	if err != nil {
	// 		log.Errorw("check dir is empty", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}

	// 	// 如果目录不为空，删除并重建目录
	// 	if !empty {
	// 		if err := removeDir(dir); err != nil {
	// 			log.Errorw("remove dir", "project", p.Name, "err", err)
	// 			return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 		}
	// 		if err := mkDir(dir); err != nil {
	// 			log.Errorw("create dir", "project", p.Name, "err", err)
	// 			return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 		}
	// 	}

	// 	// 克隆项目
	// 	gitRepo, err = git.PlainClone(dir, false, &git.CloneOptions{
	// 		Auth: publicKeys,
	// 		URL:  repo,
	// 	})

	// 	if err != nil {
	// 		log.Errorw("clone project", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}

	// }

	// // 检查 master 分支是不是最新的
	// // git ls-remote origin master
	// // git.NewRemote()

	// hl, err := gitRepo.ResolveRevision(plumbing.Revision("master"))
	// if err != nil {
	// 	log.Errorw("resolve revision failed", "project", p.Name, "err", err)
	// 	return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// }
	// remote, err := gitRepo.Remote("origin")
	// if err != nil {
	// 	log.Errorw("get remote failed", "project", p.Name, "err", err)
	// 	return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// }
	// refs, err := remote.List(&git.ListOptions{
	// 	Auth: publicKeys,
	// })
	// if err != nil {
	// 	log.Errorw("list remote failed", "project", p.Name, "err", err)
	// 	return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// }

	// var hr string
	// for _, ref := range refs {
	// 	if ref.Name().String() == "refs/heads/master" {
	// 		hr = ref.Hash().String()
	// 	}
	// }

	// var needUpdate bool

	// if hr != "" && hl.String() != hr {
	// 	needUpdate = true
	// }

	// // 如果需要更新代码，则执行 git pull
	// if needUpdate {
	// 	needBuild = true
	// 	log.Infow("project is out of date, need rebuild", "project", p.Name)
	// 	w, err := gitRepo.Worktree()
	// 	if err != nil {
	// 		log.Errorw("get worktree failed", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}
	// 	err = w.Pull(&git.PullOptions{
	// 		RemoteName: "origin",
	// 	})
	// 	if err != nil {
	// 		log.Errorw("pull remote repo", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}
	// } else {
	// 	log.Debugw("project is up to date", "project", p.Name)
	// }

	// ref, err := gitRepo.Head()
	// if err != nil {
	// 	log.Errorw("get head failed", "project", p.Name, "err", err)
	// 	return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// }
	// log.Debugw("get head", "project", p.Name, "ref", ref)
	// commit, err := gitRepo.CommitObject(ref.Hash())
	// if err != nil {
	// 	log.Errorw("get commit failed", "project", p.Name, "err", err)
	// 	return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// }

	// log.Debugw("get commit", "project", p.Name, "commit", commit)

	// if needBuild {
	// 	if output, err = installDeps(dir); err != nil {
	// 		log.Errorw("install dependencies failed", "project", p.Name, "err", err, "output", string(output))
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}

	// 	if output, err = buildProject(dir, p.BuildCmd); err != nil {
	// 		log.Errorw("build project failed", "project", p.Name, "err", err, "output", string(output))
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}

	// 	// 拷贝打包出的文件至 www 下
	// 	src := fmt.Sprintf("%s/%s", dir, p.Dist)
	// 	dst := fmt.Sprintf("%s/%s", www, p.Name)

	// 	if err := copyDir(src, dst); err != nil {
	// 		log.Errorw("copy dir failed", "project", p.Name, "err", err)
	// 		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	// 	}
	// }

	// return location, sql, nil
}

// checkDir 检查目录是否存在，不存在则创建
func checkDir(home string) error {
	if _, err := os.Stat(home); os.IsNotExist(err) {
		if err := os.MkdirAll(home, 0755); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	if err := removeDir(dst); err != nil {
		return err
	}
	if err := mkDir(dst); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Generate the destination path
		destPath := filepath.Join(dst, path[len(src):])

		if info.IsDir() {
			// Create the directory in the destination
			err := os.MkdirAll(destPath, 0755)
			if err != nil {
				return err
			}
		} else {
			// Copy the file to the destination
			err := copyFile(path, destPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func copyFile(src, dst string) error {
	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Copy the file mode
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

func removeFile(path string) error {
	// check if the file is existed
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return os.Remove(path)
	}
	return nil
}

func compressDir(src string, dst string) error {
	// 检查文件是否存在
	_, err := os.Stat(dst)

	// 存在则先删除
	if os.IsExist(err) {
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	zipFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if path == src {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, src+"/")
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)

		return err

	})

}

func isDirExist(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.ReadDir(1)
	if err == nil {
		return false, nil
	}

	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// removeDir 删除目录
func removeDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	} else {
		return os.RemoveAll(dir)
	}
}

// mkDir 创建目录
func mkDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func cloneRepo(repo string, dir string) (string, error) {
	cmd := exec.Command("git", "clone", repo, dir)
	cmd.Dir = dir
	return execCommand(cmd)
}

func installDeps(dir string) (string, error) {
	cmd := exec.Command("yarn", "install")
	cmd.Dir = dir
	return execCommand(cmd)
}

func buildProject(dir string, buildCmd string) (string, error) {
	cmds := strings.Split(buildCmd, " ")
	if len(cmds) < 2 {
		return "", fmt.Errorf("wrong build command: %s", buildCmd)
	}
	if cmds[0] != "yarn" {
		return "", fmt.Errorf("wrong build command: %s", buildCmd)
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Dir = dir
	return execCommand(cmd)
}

func formatDuration(duration time.Duration) string {

	days := int(duration.Hours()) / 24      // calculate the number of days from hours
	hours := int(duration.Hours()) % 24     // calculate the remaining hours
	minutes := int(duration.Minutes()) % 60 // calculate the remaining minutes
	seconds := int(duration.Seconds()) % 60 // calculate the remaining seconds

	dataString := fmt.Sprintf("%02d:%02d:%02d:%02d", days, hours, minutes, seconds) // f7325ormat the output string

	var output []string

	zeroRex := regexp.MustCompile(`^0+$`)

	for _, section := range strings.Split(dataString, ":") {
		if !zeroRex.MatchString(section) {
			output = append(output, section)
		} else if len(output) > 0 {
			output = append(output, section)
		}
	}

	units := []string{"秒", "分钟", "小时", "天"}

	// reverse the output
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}

	length := len(output)

	for i := 0; i < length; i++ {
		output = append(output[:(0+i*2)], append([]string{units[i]}, output[(i*2):]...)...)
	}

	// reverse the output array
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}

	return strings.Join(output, "")
}

func checkBucket(ctx context.Context, bucket string) error {
	exists, err := storage.S.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := storage.S.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func uploadFiles(ctx context.Context, bucketName, objectPrefix, filePath string) error {
	return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		objectName := objectPrefix + strings.TrimPrefix(path, filePath)

		if _, err := storage.S.FPutObject(ctx, bucketName, objectName, path, minio.PutObjectOptions{}); err != nil {
			return err
		}
		return nil
	})

	//if info, err := storage.S.FPutObject(ctx, bucketName, objectPrefix, filePath, minio.PutObjectOptions{}); err != nil {
	//	return err
	//} else {
	//	log.Infow("upload file success", "info", info)
	//}
	//
	//return nil
}

func cleanYarnCache() {
	cmd := exec.Command("yarn", "cache", "clean")
	log.Infow("exec command", "cmd", cmd.String())
	cmd.Output()
}

func execCommand(cmd *exec.Cmd) (string, error) {
	log.Infow("exec command", "cmd", cmd.String())
	msg, err := cmd.CombinedOutput()

	return fmt.Sprintf("output: %s", msg), err

}
