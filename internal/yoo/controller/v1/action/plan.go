package action

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"

	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	"phos.cc/yoo/internal/pkg/log"
	veldt "phos.cc/yoo/internal/pkg/validator"
	"phos.cc/yoo/internal/yoo/biz"
	"phos.cc/yoo/internal/yoo/storage"
	v1 "phos.cc/yoo/pkg/api/yoo/v1"
)

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

func (ctrl *ActionController) Download(c *gin.Context) {}

func execPlan(c context.Context, b biz.Biz, pid int32, gid int32, isMicro bool) {
	log.C(c).Infow("exec plan", "pid", pid, "gid", gid)

	// 检查缓存目录
	log.Infow("check home dir", "home", "/tmp/.yoo")
	// 2. 检查缓存目录
	home := fmt.Sprintf("/tmp/.yoo/project-%d", pid)
	if err := checkDir(home); err != nil {
		log.Errorw("check home dir", "err", err)
		return
	}

	assets := fmt.Sprintf("%s/assets", home)

	// 检查 assets 目录是否存在
	if err := checkDir(assets); err != nil {
		log.Errorw("check assets dir", "err", err)
		return
	}

	// 检查 bundles 目录是否存在
	bundles := fmt.Sprintf("%s/bundles", home)
	if err := checkDir(bundles); err != nil {
		log.Errorw("check bundles dir", "err", err)
		return
	}

	page := 1
	const pageSize = 10
	r := &v1.ListTaskQuery{
		Page:     page,
		PageSize: pageSize,
		PlanID:   int(pid),
	}
	// TODO: 查询 group
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

	for int64((page-1)*pageSize) <= total {
		var list []*v1.GetTaskResponse
		var err error
		list, total, err = b.Tasks().List(c, r)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		page += 1
		g, ctx := errgroup.WithContext(c)
		for _, task := range list {
			g.Go(func() error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					log.Infow("query project", "id", task.ProjectID)
					// 1. 查询 project
					p, err := b.Projects().Get(c, task.ProjectID)
					if err != nil {
						return err
					}

					log.Infow("check project dir", "dir", fmt.Sprintf("%s/%s", home, p.Name))

					dir := fmt.Sprintf("%s/%s", home, p.Name)

					// 3. 检查项目目录
					if err := checkDir(dir); err != nil {
						return err
					}

					// 4. 清空项目目录
					if err := cleanupDir(dir); err != nil {
						return err
					}

					// 5. 克隆项目到本地
					var repo string
					if p.SSHURL != "" {
						repo = p.SSHURL
					} else {
						repo = p.HTTPURL
					}
					var output []byte
					if output, err = cloneRepo(repo, dir); err != nil {
						log.Errorw("clone project", "err", err, "output", string(output))
						return err
					}
					log.Infow("clone project success", "output", string(output))

					// 6. 安装项目依赖
					if output, err = installDeps(dir); err != nil {
						log.Errorw("install dependencies", "err", err, "output", string(output))
						return err
					}
					log.Infow("install dependencies success", "output", string(output))

					// 7. 执行打包命令
					if output, err = buildProject(dir, p.BuildCmd); err != nil {
						log.Errorw("build project", "err", err, "output", string(output))
						return err
					}
					log.Infow("build success", "output", string(output))

					// 拷贝打包出的文件至 assets 下
					src := fmt.Sprintf("%s/%s", dir, p.Dist)
					dst := fmt.Sprintf("%s/%s", assets, p.Name)
					if err := copyDir(src, dst); err != nil {
						log.Errorw("copy dir", "err", err)
						return err
					}
					log.Infow("copy dir success", "src", src, "dst", dst)

					nginxsb.WriteString(fmt.Sprintf(`
	location /%s/{
	    alias /usr/share/nginx/www/%s/;
	    index index.html index.htm;
	    try_files $uri /%s/index.html;
    }
`, p.Name, p.Name, p.Name))

					sqlsb.WriteString(fmt.Sprintf("INSERT INTO `resources` (`name`, `label`, `badge`, `created_at`, `updated_at`) VALUES ('%s', '%s', '%s', NOW(), NOW())", p.Name, p.Description, ""))

					//TODO: 8. 检查存储桶
					//bucketName := fmt.Sprintf("yoo")
					//if err := checkBucket(c, bucketName); err != nil {
					//	return err
					//}

					//TODO: 9. 上传文件
					//dist := fmt.Sprintf("%s/%s", dir, p.Dist)
					//if err := uploadFiles(c, bucketName, p.Name, dist); err != nil {
					//	return err
					//}

					log.Infow("task done")
					return nil
				}
			})
		}
		if err := g.Wait(); err != nil {
			log.Errorw("exec plan error", "err", err)
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

	// 将 assets 目录下的文件压缩到 bundles 目录下
	if err := compressDir(assets, fmt.Sprintf("%s/assets.zip", bundles)); err != nil {
		log.Errorw("compress dir", "err", err)
		return
	}

	// 在 bundles 目录下生成 nginx.conf
	file, err := os.Create(fmt.Sprintf("%s/nginx.conf", bundles))
	if err != nil {
		log.Errorw("create nginx.conf error", "err", err)
		return
	}
	defer file.Close()
	if _, err := file.WriteString(nginxsb.String()); err != nil {
		log.Errorw("write nginx.conf error", "err", err)
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
	if err := checkDir(dst); err != nil {
		return err
	}
	if err := cleanupDir(dst); err != nil {
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

// cleanupDir 清空目录，可在第一级进行忽略
func cleanupDir(dir string, excludes ...string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
outer:
	for _, file := range files {
		if len(excludes) > 0 {
			for _, exclude := range excludes {
				if file == exclude {
					continue outer
				}
			}
		}
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}
	return nil
}

func cloneRepo(repo string, dir string) ([]byte, error) {
	cmd := exec.Command("git", "clone", repo, dir)
	cmd.Dir = dir
	return cmd.Output()
}

func installDeps(dir string) ([]byte, error) {
	cmd := exec.Command("yarn", "install", "--registry=https://registry.npm.taobao.org")
	cmd.Dir = dir
	return cmd.Output()
}

func buildProject(dir string, buildCmd string) ([]byte, error) {
	cmds := strings.Split(buildCmd, " ")
	if len(cmds) < 2 {
		return nil, fmt.Errorf("wrong build command: %s", buildCmd)
	}
	if cmds[0] != "yarn" {
		return nil, fmt.Errorf("wrong build command: %s", buildCmd)
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Dir = dir
	return cmd.Output()
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
