package api

import (
	"game/protocol"
	"game/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RoomHandler 定义房间 API 处理函数结构
type RoomHandler struct {
	roomService service.RoomService
}

// NewRoomHandler 创建 RoomHandler 实例
func NewRoomHandler(roomService service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// CreateRoom 处理创建房间请求
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req protocol.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 从请求中获取用户名（这里简化处理，实际应该从认证信息中获取）
	// 注意：这里需要从其他地方获取用户名，比如查询参数或认证信息
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "用户名不能为空",
		})
		return
	}

	// 调用 Service 层处理创建房间逻辑
	room, err := h.roomService.CreateRoom(req, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, protocol.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "创建房间失败: " + err.Error(),
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, protocol.CreateRoomResponse{
		Success: true,
		Message: "房间创建成功",
		RoomID:  room.ID,
	})
}

// JoinRoom 处理加入房间请求
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	var req protocol.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 从请求中获取用户名（这里简化处理，实际应该从认证信息中获取）
	// 注意：这里需要从其他地方获取用户名，比如查询参数或认证信息
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "用户名不能为空",
		})
		return
	}

	// 调用 Service 层处理加入房间逻辑
	room, message, err := h.roomService.JoinRoom(req, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, protocol.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "加入房间失败: " + err.Error(),
		})
		return
	}

	if room == nil {
		c.JSON(http.StatusOK, protocol.JoinRoomResponse{
			Success: false,
			Message: message,
		})
		return
	}

	// 构建房间信息响应
	roomInfo := protocol.RoomInfo{
		ID:         room.ID,
		Name:       room.Name,
		Host:       room.HostID,
		Players:    room.Players,
		MaxPlayers: room.MaxPlayers,
		Status:     room.Status,
	}

	// 返回响应
	c.JSON(http.StatusOK, protocol.JoinRoomResponse{
		Success: true,
		Message: message,
		Room:    roomInfo,
	})
}

// GetRoomList 处理获取房间列表请求
func (h *RoomHandler) GetRoomList(c *gin.Context) {
	// 调用 Service 层获取所有房间
	rooms := h.roomService.GetAllRooms()

	// 构建房间列表响应
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

	// 返回响应
	c.JSON(http.StatusOK, protocol.RoomListResponse{
		Rooms: roomInfos,
	})
}