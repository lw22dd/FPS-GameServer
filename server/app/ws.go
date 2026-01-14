package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"game/crypto"
	"game/data"
	"game/models"
	"game/protocol"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket 相关变量
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

// Hub 定义 WebSocket 中心结构，这里就是WS服务端
type Hub struct {
	clients      map[*Client]bool // 这里存储所有活跃的客户端
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

				// 更新用户状态：离线，清除房间ID
				user := h.userStore.FindByUsername(client.username)
				if user != nil && user.Online {
					user.Online = false
					user.RoomID = ""
					h.userStore.Update(client.username, *user)
					log.Printf("用户 %s 断开连接，已更新状态为离线", client.username)
				}
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

					// 更新用户状态：离线，清除房间ID
					user := h.userStore.FindByUsername(client.username)
					if user != nil && user.Online {
						user.Online = false
						user.RoomID = ""
						h.userStore.Update(client.username, *user)
						log.Printf("用户 %s 连接超时，已更新状态为离线", client.username)
					}
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
					log.Printf("用户 %s 心跳超时，已自动下线", username)
				}
				delete(h.heartbeatMap, username)
			}
		}
		h.mu.Unlock()
	}
}

// HasActiveConnection 检查指定用户名是否存在活跃的WebSocket连接
func (h *Hub) HasActiveConnection(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 检查心跳映射中是否存在该用户
	_, exists := h.heartbeatMap[username]
	if exists {
		log.Printf("检测到用户 %s 已存在活跃的WebSocket连接", username)
		return true
	}

	// 同时检查clients map中是否存在该用户的连接
	for client := range h.clients {
		if client.username == username {
			log.Printf("检测到用户 %s 已存在活跃的WebSocket连接", username)
			return true
		}
	}

	return false
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
		// 调用websocketAPI 读取消息
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// 解密消息
		decryptedMsg, err := crypto.Decrypt(string(message))
		if err != nil {
			log.Printf("解密消息失败: %v", err)
			continue
		}

		c.hub.handleMessage(c, []byte(decryptedMsg))
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(2 * time.Second) // 心跳间隔，默认每2秒发送一次心跳
	defer func() {
		ticker.Stop()
		c.conn.Close()
		// 连接关闭时，通知Hub注销客户端
		c.hub.unregister <- c
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 对称加密消息
			encryptedMsg, err := crypto.Encrypt(string(message))
			if err != nil {
				log.Printf("加密消息失败: %v", err)
				continue
			}

			c.conn.WriteMessage(websocket.TextMessage, []byte(encryptedMsg))

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
		h.startGame(client) // 对房主所在的客户端启动游戏

	// 创建房间管理相关消息处理
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
		var joinReq protocol.JoinRoomRequest //获取前端发送的加入房间ID
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
			user.RoomID = room.ID // 更新用户所在房间ID
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
	room := h.roomStore.GetByID(client.roomID) // 从房间存储中获取房间，在房间中开始游戏
	if room == nil {
		return
	}

	if client.username != room.HostID {
		return
	}

	room.Status = "playing"
	h.roomStore.Update(*room)

	gameStart := protocol.Message{ // 游戏开始消息，准备广播
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
	for c := range h.clients { // 遍历所有客户端并发送开始游戏消息
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

	// 1. 检查该用户是否已经有活跃的WebSocket连接，用于检查用户已登录
	if s.hub.HasActiveConnection(username) {
		log.Printf("拒绝重复连接: 用户 %s 已存在活跃的WebSocket连接", username)
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户已登录"})
		return
	}

	// 2. 检查用户是否已登录
	user := s.userStore.FindByUsername(username)
	if user == nil || !user.Online {
		log.Printf("拒绝未登录连接: 用户 %s 未登录或不存在", username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
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
		roomID:   user.RoomID,
	}

	log.Printf("用户 %s 建立WebSocket连接成功", username)
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
