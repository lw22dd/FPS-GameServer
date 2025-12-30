# 双人联机实现指南

## 一、双人联机整体流程

双人联机游戏的核心是通过服务器作为中介，实现两个客户端之间的实时通信和状态同步。以下是完整的实现流程：

```
客户端1 <--> 游戏服务器 <--> 客户端2
```

## 二、实现步骤

### 1. 准备工作

#### 1.1 服务器准备
- 确保游戏服务器已经启动，运行在 `http://localhost:8080`
- WebSocket服务地址：`ws://localhost:8080/ws`

#### 1.2 客户端准备
- 创建两个Vue客户端实例（可以在两个浏览器标签页中打开同一应用）
- 确保客户端能够访问服务器地址

### 2. 客户端连接流程

#### 2.1 注册/登录
1. 客户端1和客户端2分别注册或登录账号
2. 服务器验证用户信息，返回登录结果
3. 登录成功后，客户端获取房间列表

#### 2.2 房间系统
1. 客户端1创建房间，成为房主
2. 服务器创建房间，返回房间信息
3. 客户端2获取房间列表，看到客户端1创建的房间
4. 客户端2加入该房间
5. 服务器验证房间状态，添加玩家到房间
6. 服务器通知房间内所有玩家，房间状态更新

#### 2.3 游戏开始
1. 房主（客户端1）点击开始游戏按钮
2. 服务器收到开始游戏请求，验证房主权限
3. 服务器更新房间状态为"playing"
4. 服务器向房间内所有玩家发送游戏开始通知
5. 客户端1和客户端2收到通知，进入游戏场景

### 3. 游戏状态同步

#### 3.1 WebSocket连接建立
- 客户端进入游戏场景后，建立WebSocket连接
- 连接URL格式：`ws://localhost:8080/ws?username=xxx`
- 服务器为每个客户端创建一个Client实例，管理连接

#### 3.2 初始游戏状态
- 服务器初始化两个英雄（hero1和hero2）的初始位置
- 服务器向所有客户端发送初始游戏状态
- 客户端根据初始状态渲染游戏场景

#### 3.3 实时状态同步
- **位置同步**：玩家移动时，客户端发送位置更新消息
  ```json
  {
    "type": "player_action",
    "payload": {
      "player_id": "player1",
      "action": "move",
      "x": 100,
      "y": 200,
      "direction": 1
    }
  }
  ```
- **朝向同步**：玩家改变朝向时，客户端发送朝向更新消息
- **开火同步**：玩家开火时，客户端发送开火消息，服务器广播给所有玩家
- **命中同步**：子弹击中英雄时，客户端发送命中消息，服务器更新血量并广播
- **死亡同步**：英雄血量为0时，客户端发送死亡消息，服务器处理游戏结束

### 4. 心跳机制

- 客户端每2秒发送一次心跳请求
  ```json
  {"type": "heartbeat"}
  ```
- 服务器收到心跳后更新时间戳
- 超过10秒未收到心跳，服务器标记玩家离线
- 游戏过程中心跳持续发送，确保连接稳定

### 5. 游戏结束

- 当一个英雄死亡时，客户端发送死亡消息
- 服务器判定游戏结束，记录游戏结果
- 服务器向所有玩家广播游戏结束通知
  ```json
  {
    "type": "game_over",
    "payload": {
      "winner": "player1",
      "loser": "player2",
      "duration": 120
    }
  }
  ```
- 客户端收到通知，显示游戏结果

## 三、关键代码实现

### 1. 服务器端WebSocket处理

```go
// main.go - WebSocket连接处理
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }

    username := r.URL.Query().Get("username")
    if username == "" {
        conn.Close()
        return
    }

    client := &Client{
        hub:      hub,
        conn:     conn,
        send:     make(chan []byte, 256),
        username: username,
        roomID:   "",
    }

    // 设置房间ID
    user := hub.userStore.FindByUsername(username)
    if user != nil {
        client.roomID = user.RoomID
    }

    hub.register <- client

    // 启动读写协程
    go client.writePump()
    go client.readPump()
}
```

### 2. 房间加入处理

```go
// room_handler.go - 加入房间
func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
    // ... 验证请求
    
    // 更新房间玩家列表
    room.Players = append(room.Players, username)
    room.Status = "ready"
    h.roomStore.Update(*room)
    
    // 更新用户房间信息
    user := h.userStore.FindByUsername(username)
    if user != nil {
        user.RoomID = room.ID
        h.userStore.Update(username, *user)
    }
    
    // 返回加入结果
    h.sendJoinRoomResponse(w, true, "加入成功", roomInfo)
}
```

### 3. 游戏状态同步

```go
// main.go - 广播游戏动作
func (h *Hub) broadcastGameAction(sender *Client, msg protocol.Message) {
    h.mu.RLock()
    for client := range h.clients {
        // 只向同一房间的其他玩家发送消息
        if client.roomID == sender.roomID && client.username != sender.username {
            data, _ := json.Marshal(msg)
            client.send <- data
        }
    }
    h.mu.RUnlock()
}
```

### 4. 客户端WebSocket连接

```javascript
// 在ShootingGame.vue中建立WebSocket连接
connectWebSocket() {
  const wsUrl = `ws://localhost:8080/ws?username=${this.username}`;
  this.ws = new WebSocket(wsUrl);
  
  this.ws.onopen = () => {
    console.log('WebSocket连接已建立');
    this.startHeartbeat();
  };
  
  this.ws.onmessage = (event) => {
    this.handleWebSocketMessage(event.data);
  };
  
  this.ws.onclose = () => {
    console.log('WebSocket连接已关闭');
    this.stopHeartbeat();
  };
}
```

### 5. 心跳机制实现

```javascript
// 客户端心跳发送
startHeartbeat() {
  this.heartbeatInterval = setInterval(() => {
    if (this.ws.readyState === WebSocket.OPEN) {
      const heartbeatMsg = { type: 'heartbeat' };
      this.ws.send(JSON.stringify(heartbeatMsg));
    }
  }, 2000); // 每2秒发送一次心跳
}
```

## 四、测试步骤

1. **启动服务器**：
   ```bash
   cd d:\lwdd\code\tmp\socket\game\server
   go run main.go
   ```

2. **启动客户端**：
   ```bash
   cd d:\lwdd\code\tmp\socket\game\shot
   npm run dev
   ```

3. **测试流程**：
   - 在浏览器中打开两个标签页，访问客户端地址（如 `http://localhost:5173`）
   - 标签页1：注册/登录账号（如 user1），创建房间
   - 标签页2：注册/登录账号（如 user2），加入房间
   - 标签页1：点击开始游戏
   - 两个标签页同时进入游戏场景，开始双人对战
   - 测试移动、开火、命中、死亡等游戏机制
   - 验证游戏结束后结果显示

## 五、关键技术点

### 1. 状态同步
- 采用"客户端预测+服务器权威"模式
- 客户端发送动作，服务器验证并广播结果
- 所有游戏状态以服务器为准

### 2. 消息序列化
- 使用JSON格式序列化消息
- 支持加密传输（AES-256）
- 消息类型清晰，便于扩展

### 3. 并发处理
- 服务器使用Goroutine处理每个连接
- 客户端使用事件驱动处理消息
- 避免阻塞操作，确保实时性

### 4. 错误处理
- 优雅处理WebSocket连接断开
- 服务器定期清理无效连接
- 客户端自动重连机制（可扩展）

## 六、扩展功能

### 1. 多人游戏
- 扩展房间最大玩家数
- 增加玩家选择角色功能
- 调整游戏场景以适应更多玩家

### 2. 游戏模式
- 团队对战模式
- 生存模式
- 闯关模式

### 3. 优化体验
- 添加游戏匹配系统
- 增加游戏内聊天功能
- 优化网络延迟处理

## 七、常见问题及解决方案

### 1. 连接失败
- 检查服务器是否正在运行
- 检查客户端网络连接
- 验证WebSocket URL是否正确

### 2. 状态不同步
- 检查消息发送和接收逻辑
- 验证服务器广播机制
- 检查客户端状态更新逻辑

### 3. 游戏卡顿
- 优化客户端渲染性能
- 减少不必要的消息发送
- 优化服务器广播效率

### 4. 数据丢失
- 实现消息确认机制
- 添加重发逻辑
- 优化网络传输

## 八、总结

双人联机游戏的实现核心是服务器中介通信和状态同步。通过WebSocket协议实现实时通信，结合房间系统管理玩家，使用消息广播机制同步游戏状态，就可以实现流畅的双人对战体验。

本指南结合现有代码结构，详细说明了双人联机的实现步骤和关键技术点，便于开发者理解和实践。在实际开发中，可以根据需求扩展更多功能，优化游戏体验。