package routers

import (
	"go_chat_v1_test/internal/container"

	"go_chat_v1_test/internal/middlewares"
	"go_chat_v1_test/internal/validator"

	"github.com/gin-gonic/gin"
)

func SetupRouter(c *container.Container) *gin.Engine {
	router := gin.New()
	router.LoadHTMLGlob("*.html")
	// 应用中间件
	router.Use(middlewares.Recover())
	router.Use(middlewares.LanguageMiddleware())
	//
	// router.POST("/test-login", func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{
	// 		"message": "test-login works",
	// 		"path":    ctx.Request.URL.Path,
	// 	})
	// })
	// 创建验证器
	v := validator.NewValidator()
	// API 版本分组
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// 公开路由（无需认证）
			v1.POST("/auth/login", c.AuthController.Login())
			v1.POST("/auth/refresh", c.AuthController.Refresh())
			v1.POST("/auth/logout", c.AuthController.Logout())
			v1.POST("/common/sendPhoneCode", c.CommonController.SendPhoneCode(v))
			v1.POST("/creatUser", c.UserController.CreateUser(v))
			v1.POST("/common/getUserList2", c.UserController.GetAllUserByPage2(v))
			v1.POST("/user/pubMsg", c.UserController.PubMessage())
			v1.POST("/user/subMsg", c.UserController.SubMessage())
		}
	}
	//首页
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "home.html", nil)
	})

	router.GET("/ws", c.WebSocketController.Getws())
	// 健康检查
	router.GET("/health", c.HealthController.Health())

	// 静态文件服务
	router.Static("/static", "./static")

	// 404 处理
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"error": "not found"})
	})
	// // 在返回前打印所有路由（用于调试）
	// for _, route := range router.Routes() {
	// 	fmt.Printf("Route: %s %s\n", route.Method, route.Path)
	// }
	return router
}
