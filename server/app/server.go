package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"game/api"
	"game/data"
	"game/models"
	"game/protocol"
	"game/repository"
	"game/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	// 初始化数据存储
	userStore := data.NewUserStore()
	roomStore := data.NewRoomStore()
	resultStore := data.NewResultStore()

	// 初始化仓库
	userRepo := repository.NewUserRepository(userStore)
	roomRepo := repository.NewRoomRepository(roomStore)
	resultRepo := repository.NewResultRepository(resultStore)

	// 初始化服务
	userService := service.NewUserService(userRepo)
	roomService := service.NewRoomService(roomRepo, userRepo, resultRepo)

	// 初始化路由器
	router := api.NewRouter(userService, roomService)

	// 初始化 Hub
	hub := newHub(userStore, roomStore, resultStore)

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

	// 添加 WebSocket 路由
	s.router.Engine.GET("/ws", s.serveWs)

	// 启动 Hub
	go s.hub.run()
	go s.hub.heartbeatCheck()

	// 启动 HTTP 服务器
	log.Println("游戏服务器启动在 http://localhost:8080")
	log.Println("WebSocket: ws://localhost:8080/ws?username=xxx")
	return s.router.Run(":8080")
}

// 以下是 WebSocket 相关代码，保留原有功能

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client 定义 WebSocket 客户端结构
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	username string
	roomID   string
	lastPing time.Time
}

// Hub 定义 WebSocket 中心结构
type Hub struct {
	clients      map[*Client]bool
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
	userStore    *data.UserStore
	roomStore    *data.RoomStore
	resultStore  *data.ResultStore
	mu           sync.RWMutex
	heartbeatMap map[string]time.Time
}

// newHub 创建 Hub 实例
func newHub(userStore *data.UserStore, roomStore *data.RoomStore, resultStore *data.ResultStore) *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		userStore:    userStore,
		roomStore:    roomStore,
		resultStore:  resultStore,
		heartbeatMap: make(map[string]time.Time),
	}
}

// run 运行 Hub
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.heartbeatMap[client.username] = time.Now()
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.heartbeatMap, client.username)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					delete(h.heartbeatMap, client.username)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// heartbeatCheck 检查心跳
func (h *Hub) heartbeatCheck() {
	for {
		time.Sleep(1 * time.Second)
		h.mu.Lock()
		now := time.Now()
		for username, lastPing := range h.heartbeatMap {
			if now.Sub(lastPing) > 10*time.Second {
				user := h.userStore.FindByUsername(username)
				if user != nil && user.Online {
					user.Online = false
					user.RoomID = ""
					h.userStore.Update(username, *user)
				}
				delete(h.heartbeatMap, username)
			}
		}
		h.mu.Unlock()
	}
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.hub.mu.Lock()
		c.hub.heartbeatMap[c.username] = time.Now()
		c.hub.mu.Unlock()
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.hub.handleMessage(c, message)
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(2 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (h *Hub) handleMessage(client *Client, message []byte) {
	var msg protocol.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	switch msg.Type {
	case protocol.MsgTypeHeartbeat:
		h.mu.Lock()
		h.heartbeatMap[client.username] = time.Now()
		h.mu.Unlock()
		reply := protocol.Message{Type: protocol.MsgTypeHeartbeatReply}
		data, _ := json.Marshal(reply)
		client.send <- data

	case protocol.MsgTypePlayerAction:
		var action protocol.PlayerAction
		json.Unmarshal(msg.Payload, &action)
		h.broadcastGameAction(client, msg)

	case protocol.MsgTypeFire:
		var fire protocol.FireAction
		json.Unmarshal(msg.Payload, &fire)
		h.broadcastGameAction(client, msg)

	case protocol.MsgTypeHit:
		var hit protocol.HitAction
		json.Unmarshal(msg.Payload, &hit)
		h.broadcastGameAction(client, msg)

	case protocol.MsgTypeDeath:
		var death struct {
			PlayerID string `json:"player_id"`
		}
		json.Unmarshal(msg.Payload, &death)
		h.handleDeath(client, death.PlayerID)

	case protocol.MsgTypeGameOver:
		var gameOver protocol.GameOverInfo
		json.Unmarshal(msg.Payload, &gameOver)
		h.handleGameOver(gameOver)

	case protocol.MsgTypeStartGame:
		h.startGame(client)

	// 添加房间管理相关消息处理
	case protocol.MsgTypeCreateRoom:
		var createReq protocol.CreateRoomRequest
		if err := json.Unmarshal(msg.Payload, &createReq); err != nil {
			break
		}

		room := models.Room{
			ID:         fmt.Sprintf("room_%d", time.Now().UnixNano()),
			Name:       createReq.Name,
			HostID:     client.username,
			Players:    []string{client.username},
			MaxPlayers: createReq.MaxPlayers,
			Status:     "waiting",
			CreatedAt:  time.Now(),
		}
		h.roomStore.Add(room)

		client.roomID = room.ID
		user := h.userStore.FindByUsername(client.username)
		if user != nil {
			user.RoomID = room.ID
			h.userStore.Update(client.username, *user)
		}

		// 返回房间信息给客户端
		roomInfo := protocol.RoomInfo{
			ID:         room.ID,
			Name:       room.Name,
			Host:       room.HostID,
			Players:    room.Players,
			MaxPlayers: room.MaxPlayers,
			Status:     room.Status,
		}
		respMsg := protocol.Message{
			Type: protocol.MsgTypeJoinRoomResult,
			Payload: mustMarshal(protocol.JoinRoomResponse{
				Success: true,
				Message: "房间创建成功",
				Room:    roomInfo,
			}),
		}
		respData, _ := json.Marshal(respMsg)
		client.send <- respData

	case protocol.MsgTypeRoomList:
		// 返回房间列表给客户端
		rooms := h.roomStore.GetAll()
		roomInfos := make([]protocol.RoomInfo, 0)
		for _, room := range rooms {
			if room.Status != "playing" {
				roomInfos = append(roomInfos, protocol.RoomInfo{
					ID:         room.ID,
					Name:       room.Name,
					Host:       room.HostID,
					Players:    room.Players,
					MaxPlayers: room.MaxPlayers,
					Status:     room.Status,
				})
			}
		}

		respMsg := protocol.Message{
			Type:    protocol.MsgTypeRoomList,
			Payload: mustMarshal(protocol.RoomListResponse{Rooms: roomInfos}),
		}
		respData, _ := json.Marshal(respMsg)
		client.send <- respData

	case protocol.MsgTypeJoinRoom:
		var joinReq protocol.JoinRoomRequest
		if err := json.Unmarshal(msg.Payload, &joinReq); err != nil {
			break
		}

		room := h.roomStore.GetByID(joinReq.RoomID)
		if room == nil {
			respMsg := protocol.Message{
				Type: protocol.MsgTypeJoinRoomResult,
				Payload: mustMarshal(protocol.JoinRoomResponse{
					Success: false,
					Message: "房间不存在",
				}),
			}
			respData, _ := json.Marshal(respMsg)
			client.send <- respData
			break
		}

		if room.Status == "playing" {
			respMsg := protocol.Message{
				Type: protocol.MsgTypeJoinRoomResult,
				Payload: mustMarshal(protocol.JoinRoomResponse{
					Success: false,
					Message: "游戏进行中，无法加入",
				}),
			}
			respData, _ := json.Marshal(respMsg)
			client.send <- respData
			break
		}

		if len(room.Players) >= room.MaxPlayers {
			respMsg := protocol.Message{
				Type: protocol.MsgTypeJoinRoomResult,
				Payload: mustMarshal(protocol.JoinRoomResponse{
					Success: false,
					Message: "房间已满",
				}),
			}
			respData, _ := json.Marshal(respMsg)
			client.send <- respData
			break
		}

		for _, player := range room.Players {
			if player == client.username {
				respMsg := protocol.Message{
					Type: protocol.MsgTypeJoinRoomResult,
					Payload: mustMarshal(protocol.JoinRoomResponse{
						Success: false,
						Message: "您已在房间中",
					}),
				}
				respData, _ := json.Marshal(respMsg)
				client.send <- respData
				break
			}
		}

		// 添加玩家到房间
		room.Players = append(room.Players, client.username)
		if len(room.Players) >= 2 {
			room.Status = "ready"
		}
		h.roomStore.Update(*room)

		client.roomID = room.ID
		user := h.userStore.FindByUsername(client.username)
		if user != nil {
			user.RoomID = room.ID
			h.userStore.Update(client.username, *user)
		}

		// 返回加入结果给客户端
		roomInfo := protocol.RoomInfo{
			ID:         room.ID,
			Name:       room.Name,
			Host:       room.HostID,
			Players:    room.Players,
			MaxPlayers: room.MaxPlayers,
			Status:     room.Status,
		}
		respMsg := protocol.Message{
			Type: protocol.MsgTypeJoinRoomResult,
			Payload: mustMarshal(protocol.JoinRoomResponse{
				Success: true,
				Message: "加入房间成功",
				Room:    roomInfo,
			}),
		}
		respData, _ := json.Marshal(respMsg)
		client.send <- respData

		// 向房间内的其他玩家广播房间信息更新
		broadcastMsg := protocol.Message{
			Type: protocol.MsgTypeJoinRoomResult,
			Payload: mustMarshal(protocol.JoinRoomResponse{
				Success: true,
				Message: client.username + " 加入了房间",
				Room:    roomInfo,
			}),
		}
		broadcastData, _ := json.Marshal(broadcastMsg)

		h.mu.RLock()
		for c := range h.clients {
			if c.roomID == room.ID && c.username != client.username {
				c.send <- broadcastData
			}
		}
		h.mu.RUnlock()
	}
}

// broadcastGameAction 广播游戏动作
func (h *Hub) broadcastGameAction(sender *Client, msg protocol.Message) {
	h.mu.RLock()
	for client := range h.clients {
		if client.roomID == sender.roomID && client.username != sender.username {
			data, _ := json.Marshal(msg)
			client.send <- data
		}
	}
	h.mu.RUnlock()
}

// handleDeath 处理死亡事件
func (h *Hub) handleDeath(loserClient *Client, loserID string) {
	room := h.roomStore.GetByID(loserClient.roomID)
	if room == nil || len(room.Players) < 2 {
		return
	}

	var winner string
	for _, player := range room.Players {
		if player != loserID {
			winner = player
			break
		}
	}

	gameOver := protocol.GameOverInfo{
		Winner: winner,
		Loser:  loserID,
	}

	h.handleGameOver(gameOver)
}

// handleGameOver 处理游戏结束事件
func (h *Hub) handleGameOver(gameOver protocol.GameOverInfo) {
	result := models.GameResult{
		ID:       fmt.Sprintf("result_%d", time.Now().UnixNano()),
		RoomID:   "",
		Winner:   gameOver.Winner,
		Loser:    gameOver.Loser,
		PlayTime: time.Now(),
		Duration: gameOver.Duration,
	}
	h.resultStore.Add(result)

	room := h.roomStore.GetByID("")
	if room != nil {
		room.Status = "waiting"
		h.roomStore.Update(*room)
	}

	msg := protocol.Message{
		Type:    protocol.MsgTypeGameOver,
		Payload: mustMarshal(gameOver),
	}
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

// startGame 处理开始游戏事件
func (h *Hub) startGame(client *Client) {
	room := h.roomStore.GetByID(client.roomID)
	if room == nil {
		return
	}

	if client.username != room.HostID {
		return
	}

	room.Status = "playing"
	h.roomStore.Update(*room)

	gameStart := protocol.Message{
		Type: protocol.MsgTypeGameStart,
		Payload: mustMarshal(protocol.RoomInfo{
			ID:         room.ID,
			Name:       room.Name,
			Players:    room.Players,
			MaxPlayers: room.MaxPlayers,
			Status:     "playing",
		}),
	}
	data, _ := json.Marshal(gameStart)

	h.mu.RLock()
	for c := range h.clients {
		if c.roomID == client.roomID {
			c.send <- data
		}
	}
	h.mu.RUnlock()
}

// serveWs 处理 WebSocket 连接
func (s *Server) serveWs(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名不能为空"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		hub:      s.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		username: username,
		roomID:   "",
	}

	user := s.userStore.FindByUsername(username)
	if user != nil {
		client.roomID = user.RoomID
	}

	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// mustMarshal 序列化数据，忽略错误
func mustMarshal(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}
