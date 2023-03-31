package action

import (
	"fmt"
	"strings"
	"testing"
)

func TestNginx(t *testing.T) {
	var nginxsb strings.Builder
	nginxsb.WriteString(`
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;
    absolute_redirect off;
    add_header Access-Control-Allow-Origin *;

    #access_log  /var/log/nginx/host.access.log  main;
`)

	nginxsb.WriteString(fmt.Sprintf(`
	location /%s/{
	    alias /usr/share/nginx/www/%s/;
	    index index.html index.htm;
	    try_files $uri /%s/index.html;
    }
`, "config-center", "config-center", "config-center"))

	nginxsb.WriteString(`
	#error_page  404              /404.html;

    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }

    # proxy the PHP scripts to Apache listening on 127.0.0.1:80
    #
    #location ~ \.php$ {
    #    proxy_pass   http://127.0.0.1;
    #}

    # pass the PHP scripts to FastCGI server listening on 127.0.0.1:9000
    #
    #location ~ \.php$ {
    #    root           html;
    #    fastcgi_pass   127.0.0.1:9000;
    #    fastcgi_index  index.php;
    #    fastcgi_param  SCRIPT_FILENAME  /scripts$fastcgi_script_name;
    #    include        fastcgi_params;
    #}

    # deny access to .htaccess files, if Apache's document root
    # concurs with nginx's one
    #
    #location ~ /\.ht {
    #    deny  all;
    #}
}
`)

	fmt.Println(nginxsb.String())
}

func TestMysql(t *testing.T) {
	var sqlsb strings.Builder
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

	fmt.Println(sqlsb.String())
}

func TestCleanUp(t *testing.T) {
	var dir = "test"
	err := cleanupDir(dir, "test/b")
	if err != nil {
		t.Error(err)
	}
}
