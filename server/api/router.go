package api

import (
	"game/service"

	"github.com/gin-gonic/gin"
)

// Router 定义路由器结构
type Router struct {
	Engine    *gin.Engine
	userService service.UserService
	roomService service.RoomService
}

// NewRouter 创建路由器实例
func NewRouter(userService service.UserService, roomService service.RoomService) *Router {
	return &Router{
		Engine:    gin.Default(),
		userService: userService,
		roomService: roomService,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes() {
	// 添加 CORS 中间件
	r.Engine.Use(corsMiddleware())

	// 用户相关路由
	userGroup := r.Engine.Group("/user")
	{
		userHandler := NewUserHandler(r.userService)
		userGroup.POST("/register", userHandler.Register)
		userGroup.POST("/login", userHandler.Login)
		userGroup.POST("/logout", userHandler.Logout)
		userGroup.GET("/test", userHandler.Test)
	}

	// 房间相关路由
	roomGroup := r.Engine.Group("/room")
	{
		roomHandler := NewRoomHandler(r.roomService)
		roomGroup.POST("/create", roomHandler.CreateRoom)
		roomGroup.POST("/join", roomHandler.JoinRoom)
		roomGroup.GET("/list", roomHandler.GetRoomList)
	}
}

// Run 启动服务器
func (r *Router) Run(addr string) error {
	return r.Engine.Run(addr)
}

// corsMiddleware 定义 CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}