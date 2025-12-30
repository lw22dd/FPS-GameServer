package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"game/data"
	"game/models"
	"game/protocol"
)

type RoomHandler struct {
	roomStore   *data.RoomStore
	userStore   *data.UserStore
	resultStore *data.ResultStore
}

func NewRoomHandler(roomStore *data.RoomStore, userStore *data.UserStore, resultStore *data.ResultStore) *RoomHandler {
	return &RoomHandler{
		roomStore:   roomStore,
		userStore:   userStore,
		resultStore: resultStore,
	}
}

func (h *RoomHandler) GetRoomList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	resp := protocol.RoomListResponse{Rooms: roomInfos}
	json.NewEncoder(w).Encode(resp)
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req protocol.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.Name == "" {
		h.sendError(w, http.StatusBadRequest, "房间名称不能为空")
		return
	}

	if req.MaxPlayers <= 0 {
		req.MaxPlayers = 2
	}

	var hostID string
	var username string

	if username = r.URL.Query().Get("username"); username != "" {
		user := h.userStore.FindByUsername(username)
		if user != nil {
			hostID = username
		}
	}

	if hostID == "" {
		h.sendError(w, http.StatusUnauthorized, "用户未登录")
		return
	}

	room := models.Room{
		ID:         generateRoomID(),
		Name:       req.Name,
		HostID:     hostID,
		Players:    []string{hostID},
		MaxPlayers: req.MaxPlayers,
		Status:     "waiting",
		CreatedAt:  time.Now(),
	}

	h.roomStore.Add(room)

	user := h.userStore.FindByUsername(hostID)
	if user != nil {
		user.RoomID = room.ID
		h.userStore.Update(hostID, *user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(protocol.RoomInfo{
		ID:         room.ID,
		Name:       room.Name,
		Host:       room.HostID,
		Players:    room.Players,
		MaxPlayers: room.MaxPlayers,
		Status:     room.Status,
	})
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var req protocol.JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	var username string
	if username = r.URL.Query().Get("username"); username == "" {
		h.sendError(w, http.StatusUnauthorized, "用户未登录")
		return
	}

	room := h.roomStore.GetByID(req.RoomID)
	if room == nil {
		h.sendJoinRoomResponse(w, false, "房间不存在", protocol.RoomInfo{})
		return
	}

	if room.Status == "playing" {
		h.sendJoinRoomResponse(w, false, "游戏进行中，无法加入", protocol.RoomInfo{})
		return
	}

	if len(room.Players) >= room.MaxPlayers {
		h.sendJoinRoomResponse(w, false, "房间已满", protocol.RoomInfo{})
		return
	}

	for _, player := range room.Players {
		if player == username {
			h.sendJoinRoomResponse(w, false, "您已在房间中", protocol.RoomInfo{})
			return
		}
	}

	room.Players = append(room.Players, username)
	room.Status = "ready"
	h.roomStore.Update(*room)

	user := h.userStore.FindByUsername(username)
	if user != nil {
		user.RoomID = room.ID
		h.userStore.Update(username, *user)
	}

	h.sendJoinRoomResponse(w, true, "加入成功", protocol.RoomInfo{
		ID:         room.ID,
		Name:       room.Name,
		Host:       room.HostID,
		Players:    room.Players,
		MaxPlayers: room.MaxPlayers,
		Status:     room.Status,
	})
}

func (h *RoomHandler) sendJoinRoomResponse(w http.ResponseWriter, success bool, message string, room protocol.RoomInfo) {
	w.Header().Set("Content-Type", "application/json")
	resp := protocol.JoinRoomResponse{
		Success: success,
		Message: message,
		Room:    room,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *RoomHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := protocol.ErrorResponse{
		Code:    status,
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}

func RegisterRoomRoutes(mux *http.ServeMux, roomStore *data.RoomStore, userStore *data.UserStore, resultStore *data.ResultStore) {
	handler := NewRoomHandler(roomStore, userStore, resultStore)
	mux.HandleFunc("/rooms/list", handler.GetRoomList)
	mux.HandleFunc("/rooms/create", handler.CreateRoom)
	mux.HandleFunc("/rooms/join", handler.JoinRoom)
}

func generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}

func GenerateRoomIDForTest() string {
	return generateRoomID()
}
