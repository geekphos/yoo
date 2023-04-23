package yoo

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "phos.cc/yoo/docs"
	"phos.cc/yoo/internal/pkg/core"
	"phos.cc/yoo/internal/pkg/errno"
	mw "phos.cc/yoo/internal/pkg/middleware"
	"phos.cc/yoo/internal/yoo/controller/v1/action"
	"phos.cc/yoo/internal/yoo/controller/v1/plan"
	"phos.cc/yoo/internal/yoo/controller/v1/project"
	"phos.cc/yoo/internal/yoo/controller/v1/socket"
	"phos.cc/yoo/internal/yoo/controller/v1/task"
	"phos.cc/yoo/internal/yoo/controller/v1/template"
	"phos.cc/yoo/internal/yoo/controller/v1/user"
	"phos.cc/yoo/internal/yoo/store"
)

func installRouters(g *gin.Engine) error {
	g.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})

	// 注册 /healthz handler.
	g.GET("/healthz", func(c *gin.Context) {
		core.WriteResponse(c, nil, map[string]string{"status": "ok"})
	})

	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// 创建 v1 路由分组
	v1 := g.Group("/v1")

	sc := socket.New()

	v1.GET("/ws", sc.Connect)

	{

		// 创建 users 路由分组
		uc := user.New(store.S)
		userv1 := v1.Group("/users")
		{
			userv1.POST("", uc.Create)
			userv1.POST("/login", uc.Login)
			userv1.POST("/refresh", uc.Refresh)

			userv1.Use(mw.Auth())
			userv1.PATCH("/:email/change-password", uc.ChangePassword)
			userv1.GET("/profile", uc.Profile)
		}

		// 创建 templates 路由分组
		tc := template.New(store.S)
		tempaltev1 := v1.Group("/templates")
		{
			tempaltev1.GET("/:id", tc.Get)
			tempaltev1.GET("", tc.List)

			tempaltev1.Use(mw.Auth())
			tempaltev1.POST("", tc.Create)
		}

		// 创建 projects 路由分组
		pc := project.New(store.S)
		projectv1 := v1.Group("/projects")
		{
			projectv1.GET("/:id", pc.Get)
			projectv1.GET("", pc.List)
			projectv1.GET("/categories", pc.Categories)
			projectv1.GET("/tags", pc.Tags)

			projectv1.Use(mw.Auth())
			projectv1.POST("", pc.Create)
			projectv1.PATCH("/:id", pc.Update)
		}

		// 创建 plans 路由分组
		plc := plan.New(store.S)
		planv1 := v1.Group("/plans")
		{
			planv1.GET("/:id", plc.Get)

			planv1.Use(mw.Auth())
			planv1.POST("", plc.Create)
			planv1.PATCH("/:id", plc.Update)
			planv1.GET("", plc.List)
		}

		// 创建 tasks 路由分组
		tsc := task.New(store.S)
		taskv1 := v1.Group("/tasks")
		{
			taskv1.GET("/:id", tsc.Get)

			taskv1.Use(mw.Auth())
			taskv1.POST("", tsc.Create)
			taskv1.GET("/list", tsc.List)
			taskv1.GET("/all", tsc.All)
		}

		// 创建 actions 路由分组
		ac := action.New(store.S)
		actionv1 := v1.Group("/actions")
		{
			actionv1.GET("/download/bundles/:id", ac.Download)

			actionv1.Use(mw.Auth())
			actionv1.POST("/exec/plan", ac.ExecPlan)
		}

	}

	return nil
}
