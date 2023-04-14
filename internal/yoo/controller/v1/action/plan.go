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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
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

	go func(c context.Context, b biz.Biz, pid int32, gid int32, isMicro bool) {
		execPlan(c, b, pid, gid, isMicro)
	}(c, ctrl.b, r.PlanID, r.GroupID, r.IsMicro)

	c.JSON(200, gin.H{})
}

func (ctrl *ActionController) Download(c *gin.Context) {
	pid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		core.WriteResponse(c, errno.ErrInvalidParameter, nil)
		return
	}
	// 返回给前端
	fileName := fmt.Sprintf("/tmp/.yoo/project-%d/bundles.zip", pid)
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

func execPlan(c context.Context, b biz.Biz, pid int32, gid int32, isMicro bool) {
	// 开始时间
	start := time.Now()

	log.C(c).Infow("start exec plan", "pid", pid, "gid", gid)
	email := c.Value(known.XEmailKey).(string)

	// 检查本次打包缓存目录 home
	log.Infow("create home dir", "home", "/tmp/.yoo")
	// 2. 检查缓存目录
	home := fmt.Sprintf("/tmp/.yoo/project-%d", pid)
	if err := mkdirDri(home); err != nil {
		log.Errorw("create home dir", "err", err)
		return
	}

	// 删除 bundles 目录
	bundles := fmt.Sprintf("%s/bundles", home)
	log.Infow("remove bundles dir", "bundles", bundles)
	if err := removeDir(bundles); err != nil {
		log.Errorw("remove bundles dir", "err", err)
		return
	}

	// 检查 home 下的 bundles 目录是否存在, 不存在则创建
	log.Infow("create bundles dir", "bundles", bundles)
	if err := mkdirDri(bundles); err != nil {
		log.Errorw("create bundles dir", "err", err)
		return
	}

	www := fmt.Sprintf("%s/www", bundles)
	// 清空 www 目录
	log.Infow("remove www dir", "www", www)
	if err := removeDir(www); err != nil {
		log.Errorw("remove www dir", "err", err)
		return
	}

	// 检查 bundles 目录下的 www 目录是否存在，不存在则创建
	log.Infow("create www dir", "www", www)
	if err := mkdirDri(www); err != nil {
		log.Errorw("create www dir", "err", err)
		return
	}

	// 检查 home 目录下的 repos 目录是否存在，不存在则创建
	repos := fmt.Sprintf("%s/repos", home)
	log.Infow("remove repos dir", "repos", repos)
	if err := removeDir(repos); err != nil {
		log.Errorw("remove repos dir", "err", err)
		return
	}

	// 清空 repos 目录
	log.Infow("create repos dir", "repos", repos)
	if err := mkdirDri(repos); err != nil {
		log.Errorw("create repos dir", "err", err)
		return
	}

	r := &v1.ListTaskQuery{
		Page:     1,
		PageSize: 10,
		PlanID:   int(pid),
	}

	var total int64
	total = math.MaxInt

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
`)
	sqlsb.WriteString(`
# 创建数据库
CREATE DATABASE IF NOT EXISTS yoo CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

# 使用创建的数据库
USE yoo;

CREATE TABLE IF NOT EXISTS ` + "`resources`" + `
(
    ` + "`id`" + `         int PRIMARY KEY AUTO_INCREMENT,
    ` + "`name`" + `       varchar(255) UNIQUE NOT NULL COMMENT '项目名',
    ` + "`label`" + `      varchar(255)        NOT NULL COMMENT '中文说明，一般为项目 description',
    ` + "`badge`" + `      varchar(255)        NOT NULL COMMENT '应用图标',
    ` + "`tags`" + `       json COMMENT '分类，标签',
    ` + "`created_at`" + ` timestamp           NOT NULL,
    ` + "`updated_at`" + ` timestamp           NOT NULL
);

`)

	for int64((r.Page-1)*r.PageSize) < total {
		var list []*v1.GetTaskResponse
		var err error
		list, total, err = b.Tasks().List(c, r)
		if err != nil {
			log.Errorw("list task", "err", err)
			socket_client.WriteJSON(email, gin.H{
				"event": "error",
				"data":  err.Error(),
			})
			return
		}
		r.Page += 1
		g, ctx := errgroup.WithContext(context.Background())
		for _, task := range list {
			log.Infow("query task: ", "page", r.Page-1, "task", task)
			func(task *v1.GetTaskResponse) {
				g.Go(func() error {
					select {
					case <-ctx.Done():
						// 将任务状态更新成失败
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 3}, task.ID)
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"data": gin.H{
								"id":     task.ID,
								"status": 3,
							},
						})
						return ctx.Err()
					default:
						// 将任务状态更新成 执行中
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 2}, task.ID)
						// 通过 websocket 发送消息
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"data": gin.H{
								"id":     task.ID,
								"status": 2,
							},
						})
						location, sql, err := runBuildFlow(task, b, c, repos, www)
						if err != nil {
							b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 3}, task.ID)
							socket_client.WriteJSON(email, gin.H{
								"event": "task",
								"data": gin.H{
									"id":     task.ID,
									"status": 3,
								},
							})
							return err
						}

						nginxsb.WriteString(location)

						sqlsb.WriteString(sql)

						// 将任务状态更新成成功
						b.Tasks().Update(c, &v1.UpdateTaskRequest{Status: 4}, task.ID)
						socket_client.WriteJSON(email, gin.H{
							"event": "task",
							"data": gin.H{
								"id":     task.ID,
								"status": 4,
							},
						})
						return nil
					}
				})
			}(task)
		}
		if err := g.Wait(); err != nil {
			log.Errorw("exec plan error", "err", err)
			socket_client.WriteJSON(email, gin.H{
				"event":   "project",
				"message": "exec plan error",
			})
			return
		}
	}

	nginxsb.WriteString(`
	error_page   500 502 503 504  /50x.html;

    location = /50x.html {
        root   /usr/share/nginx/html;
    }
}
`)

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
    image: "mysql:latest"
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: root
    volumes:
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

  yoo-nginx:
    image: "nginx:latest"
    restart: always
    ports:
      - "8989:80"
    volumes:
      - ./default.conf:/etc/nginx/conf.d/default.conf
      - ./www:/usr/share/nginx/www
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
	zipFile := fmt.Sprintf("/tmp/.yoo/project-%d/bundles.zip", pid)
	if err := compressDir(bundles, zipFile); err != nil {
		log.Errorw("Failed to zip the files")
		return
	}

	end := time.Now()

	elapsed := end.Sub(start)

	timeString := formatDuration(elapsed)

	log.Infow("exec plan success", "time", timeString)

	socket_client.WriteJSON(email, gin.H{
		"event":   "project",
		"message": "exec plan success",
		"elapsed": timeString,
	})
}

func runBuildFlow(task *v1.GetTaskResponse, b biz.Biz, c context.Context, repos string, www string) (string, string, error) {
	log.Infow("query project", "id", task.ProjectID)
	// 1. 查询 project
	p, err := b.Projects().Get(c, task.ProjectID)
	if err != nil {
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	dir := fmt.Sprintf("%s/%s", repos, p.Name)
	log.Infow("check project dir", "dir", dir)

	// 3. 删除项目目录
	if err := removeDir(dir); err != nil {
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	// 4. 创建项目目录
	if err := mkdirDri(dir); err != nil {
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}

	// 5. 克隆项目到本地
	var repo string
	if p.SSHURL != "" {
		repo = p.SSHURL
	} else {
		repo = p.SSHURL
	}
	var output string
	if output, err = cloneRepo(repo, dir); err != nil {
		log.Errorw("clone project", "err", err, "repo", repo, "output", string(output))
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}
	log.Infow("clone project success", "output", string(output))

	// 6. 安装项目依赖
	if output, err = installDeps(dir); err != nil {
		log.Errorw("install dependencies", "err", err, "output", string(output))
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}
	log.Infow("install dependencies success", "output", string(output))

	// 7. 执行打包命令
	if output, err = buildProject(dir, p.BuildCmd); err != nil {
		log.Errorw("build project", "err", err, "output", string(output))
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}
	log.Infow("build success", "output", string(output))

	// 拷贝打包出的文件至 www 下
	src := fmt.Sprintf("%s/%s", dir, p.Dist)
	dst := fmt.Sprintf("%s/%s", www, p.Name)
	if err := copyDir(src, dst); err != nil {
		log.Errorw("copy dir", "err", err)
		return "", "", TaskErr{TaskID: task.ID, Message: err.Error()}
	}
	log.Infow("copy dir success", "src", src, "dst", dst)

	location := fmt.Sprintf(`
	location /%s/{
	    alias /usr/share/nginx/www/%s/;
	    index index.html index.htm;
	    try_files $uri /%s/index.html;
    }
`, p.Name, p.Name, p.Name)

	sql := fmt.Sprintf("\nINSERT INTO `resources` (`name`, `label`, `badge`, `created_at`, `updated_at`) VALUES ('%s', '%s', '%s', NOW(), NOW())", p.Name, p.Description, "default.jpeg")

	return location, sql, nil
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
	if err := mkdirDri(dst); err != nil {
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

// mkdirDri 创建目录
func mkdirDri(dir string) error {
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
	// remove yarn.lock
	//if err := removeFile(filepath.Join(dir, "yarn.lock")); err != nil {
	//	return "", err
	//}
	cmd := exec.Command("yarn")
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

func execCommand(cmd *exec.Cmd) (string, error) {
	log.Infow("exec command", "cmd", cmd.String())
	msg, err := cmd.CombinedOutput()

	return fmt.Sprintf("output: %s", msg), err

}