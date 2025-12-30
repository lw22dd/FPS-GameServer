package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"game/data"
	"game/handlers"
	"game/models"
	"game/protocol"
)

func setupRoomTestStores(t *testing.T) (*data.RoomStore, *data.UserStore, *data.ResultStore) {
	roomStore := data.NewRoomStore()
	userStore := data.NewUserStore()
	resultStore := data.NewResultStore()

	// 清理房间数据
	for _, room := range roomStore.GetAll() {
		roomStore.Remove(room.ID)
	}

	return roomStore, userStore, resultStore
}

func TestGetRoomList_Empty(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	rec := httptest.NewRecorder()

	handler.GetRoomList(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp protocol.RoomListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Rooms) != 0 {
		t.Errorf("Expected 0 rooms, got %d", len(resp.Rooms))
	}
}

func TestGetRoomList_WithWaitingRooms(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
	}
	userStore.Add(user)

	room := models.Room{
		ID:         "room_123",
		Name:       "Test Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser"},
		MaxPlayers: 2,
		Status:     "waiting",
	}
	roomStore.Add(room)

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	rec := httptest.NewRecorder()

	handler.GetRoomList(rec, req)

	var resp protocol.RoomListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Rooms) != 1 {
		t.Errorf("Expected 1 room, got %d", len(resp.Rooms))
	}
	if resp.Rooms[0].Name != "Test Room" {
		t.Errorf("Expected room name 'Test Room', got '%s'", resp.Rooms[0].Name)
	}
}

func TestGetRoomList_ExcludesPlayingRooms(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
	}
	userStore.Add(user)

	waitingRoom := models.Room{
		ID:         "room_waiting",
		Name:       "Waiting Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser"},
		MaxPlayers: 2,
		Status:     "waiting",
	}
	roomStore.Add(waitingRoom)

	playingRoom := models.Room{
		ID:         "room_playing",
		Name:       "Playing Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser", "player2"},
		MaxPlayers: 2,
		Status:     "playing",
	}
	roomStore.Add(playingRoom)

	req := httptest.NewRequest("GET", "/rooms/list", nil)
	rec := httptest.NewRecorder()

	handler.GetRoomList(rec, req)

	var resp protocol.RoomListResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Rooms) != 1 {
		t.Errorf("Expected 1 room (excluding playing), got %d", len(resp.Rooms))
	}
	if resp.Rooms[0].Name != "Waiting Room" {
		t.Errorf("Expected room name 'Waiting Room', got '%s'", resp.Rooms[0].Name)
	}
}

func TestCreateRoom_Success(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
		Online:   true,
	}
	userStore.Add(user)

	reqBody := protocol.CreateRoomRequest{
		Name:       "New Room",
		MaxPlayers: 4,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/create?username=hostuser", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRoom(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var roomInfo protocol.RoomInfo
	json.Unmarshal(rec.Body.Bytes(), &roomInfo)

	if roomInfo.Name != "New Room" {
		t.Errorf("Expected room name 'New Room', got '%s'", roomInfo.Name)
	}
	if roomInfo.Host != "hostuser" {
		t.Errorf("Expected host 'hostuser', got '%s'", roomInfo.Host)
	}
	if roomInfo.MaxPlayers != 4 {
		t.Errorf("Expected max players 4, got %d", roomInfo.MaxPlayers)
	}
	if roomInfo.Status != "waiting" {
		t.Errorf("Expected status 'waiting', got '%s'", roomInfo.Status)
	}

	hostUser := userStore.FindByUsername("hostuser")
	if hostUser == nil {
		t.Fatal("Expected user to be found")
	}
	if hostUser.RoomID == "" {
		t.Error("Expected user to have room ID set")
	}
}

func TestCreateRoom_DefaultMaxPlayers(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
		Online:   true,
	}
	userStore.Add(user)

	reqBody := protocol.CreateRoomRequest{
		Name:       "New Room",
		MaxPlayers: 0,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/create?username=hostuser", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRoom(rec, req)

	var roomInfo protocol.RoomInfo
	json.Unmarshal(rec.Body.Bytes(), &roomInfo)

	if roomInfo.MaxPlayers != 2 {
		t.Errorf("Expected default max players 2, got %d", roomInfo.MaxPlayers)
	}
}

func TestCreateRoom_EmptyName(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
		Online:   true,
	}
	userStore.Add(user)

	reqBody := protocol.CreateRoomRequest{
		Name:       "",
		MaxPlayers: 2,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/create?username=hostuser", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRoom(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestCreateRoom_UserNotLoggedIn(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	reqBody := protocol.CreateRoomRequest{
		Name:       "New Room",
		MaxPlayers: 2,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateRoom(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestJoinRoom_Success(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	hostUser := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
		Online:   true,
	}
	userStore.Add(hostUser)

	playerUser := models.User{
		Username: "player2",
		Password: "password123",
		Email:    "player2@example.com",
		Online:   true,
	}
	userStore.Add(playerUser)

	room := models.Room{
		ID:         "room_123",
		Name:       "Test Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser"},
		MaxPlayers: 2,
		Status:     "waiting",
	}
	roomStore.Add(room)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "room_123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join?username=player2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	var resp protocol.JoinRoomResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("Expected success true, got false: %s", resp.Message)
	}
	if resp.Message != "加入成功" {
		t.Errorf("Expected message '加入成功', got '%s'", resp.Message)
	}
	if len(resp.Room.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(resp.Room.Players))
	}
	if resp.Room.Status != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", resp.Room.Status)
	}

	player := userStore.FindByUsername("player2")
	if player == nil {
		t.Fatal("Expected player user to be found")
	}
	if player.RoomID != "room_123" {
		t.Errorf("Expected player room ID 'room_123', got '%s'", player.RoomID)
	}
}

func TestJoinRoom_RoomNotFound(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "player2",
		Password: "password123",
		Email:    "player2@example.com",
		Online:   true,
	}
	userStore.Add(user)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "nonexistent_room",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join?username=player2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	var resp protocol.JoinRoomResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "房间不存在" {
		t.Errorf("Expected message '房间不存在', got '%s'", resp.Message)
	}
}

func TestJoinRoom_GameInProgress(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "player2",
		Password: "password123",
		Email:    "player2@example.com",
		Online:   true,
	}
	userStore.Add(user)

	room := models.Room{
		ID:         "room_123",
		Name:       "Test Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser"},
		MaxPlayers: 2,
		Status:     "playing",
	}
	roomStore.Add(room)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "room_123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join?username=player2", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	var resp protocol.JoinRoomResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "游戏进行中，无法加入" {
		t.Errorf("Expected message '游戏进行中，无法加入', got '%s'", resp.Message)
	}
}

func TestJoinRoom_RoomFull(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "player3",
		Password: "password123",
		Email:    "player3@example.com",
		Online:   true,
	}
	userStore.Add(user)

	room := models.Room{
		ID:         "room_123",
		Name:       "Test Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser", "player2"},
		MaxPlayers: 2,
		Status:     "ready",
	}
	roomStore.Add(room)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "room_123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join?username=player3", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	var resp protocol.JoinRoomResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "房间已满" {
		t.Errorf("Expected message '房间已满', got '%s'", resp.Message)
	}
}

func TestJoinRoom_AlreadyInRoom(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	user := models.User{
		Username: "hostuser",
		Password: "password123",
		Email:    "host@example.com",
		Online:   true,
	}
	userStore.Add(user)

	room := models.Room{
		ID:         "room_123",
		Name:       "Test Room",
		HostID:     "hostuser",
		Players:    []string{"hostuser"},
		MaxPlayers: 2,
		Status:     "waiting",
	}
	roomStore.Add(room)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "room_123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join?username=hostuser", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	var resp protocol.JoinRoomResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "您已在房间中" {
		t.Errorf("Expected message '您已在房间中', got '%s'", resp.Message)
	}
}

func TestJoinRoom_UserNotLoggedIn(t *testing.T) {
	roomStore, userStore, resultStore := setupRoomTestStores(t)
	handler := handlers.NewRoomHandler(roomStore, userStore, resultStore)

	reqBody := protocol.JoinRoomRequest{
		RoomID: "room_123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/rooms/join", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.JoinRoom(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestGenerateRoomID_Format(t *testing.T) {
	roomID := handlers.GenerateRoomIDForTest()

	if roomID == "" {
		t.Error("Expected room ID to be generated")
	}

	if len(roomID) <= 5 {
		t.Errorf("Expected room ID to have meaningful length, got '%s'", roomID)
	}
}
