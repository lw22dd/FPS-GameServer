package app

import (
	"log"

	"game/api"
	"game/data"
	"game/repository"
	"game/service"
)

// Server 定义服务器结构
type Server struct {
	router      *api.Router
	userStore   *data.UserStore
	roomStore   *data.RoomStore
	resultStore *data.ResultStore
	hub         *Hub
}

// NewServer 创建服务器实例
func NewServer() *Server {
	// 初始化数据存储，对应三个本地数据库
	userStore := data.NewUserStore()     //所有用户信息
	roomStore := data.NewRoomStore()     //所有房间信息
	resultStore := data.NewResultStore() //所有游戏结果信息，游戏结果不暴露给客户端

	// 初始化仓库
	userRepo := repository.NewUserRepository(userStore)
	roomRepo := repository.NewRoomRepository(roomStore)
	resultRepo := repository.NewResultRepository(resultStore)

	// 初始化服务
	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo, userRepo, resultRepo)

	// 初始化 Hub
	hub := newHub(userStore, roomStore, resultStore)

	// 初始化路由器
	router := api.NewRouter(userService, roomService)

	// 启动时的初始化清理
	log.Println("正在执行初始化清理操作...")

	// 1. 清空所有房间
	rooms := roomStore.GetAll()
	for _, room := range rooms {
		roomStore.Remove(room.ID)
	}
	log.Println("已清空所有房间")

	// 2. 重置所有用户状态（离线，清除房间ID）
	users := userStore.GetAll()
	for _, user := range users {
		if user.Online || user.RoomID != "" {
			user.Online = false
			user.RoomID = ""
			userStore.Update(user.Username, user)
		}
	}
	log.Println("已重置所有用户状态")

	return &Server{
		router:      router,
		userStore:   userStore,
		roomStore:   roomStore,
		resultStore: resultStore,
		hub:         hub,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 设置路由
	s.router.SetupRoutes()

	// 添加 WebSocket 路由，转发到 Hub
	s.router.Engine.GET("/ws", s.serveWs)

	// 启动 Hub
	go s.hub.run()
	go s.hub.heartbeatCheck()

	// 启动 HTTP 服务器
	log.Println("游戏服务器启动在 http://localhost:8080")
	log.Println("WebSocket: ws://localhost:8080/ws?username=xxx")
	return s.router.Run(":8080")
}