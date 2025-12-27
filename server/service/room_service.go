package service

import (
	"game/models"
	"game/protocol"
	"game/repository"
	"time"
)

// RoomService 定义房间业务逻辑接口
type RoomService interface {
	CreateRoom(req protocol.CreateRoomRequest, hostID string) (*models.Room, error)
	JoinRoom(req protocol.JoinRoomRequest, username string) (*models.Room, string, error)
	GetRoomByID(roomID string) *models.Room
	GetAllRooms() []models.Room
	UpdateRoom(room models.Room) bool
	RemoveRoom(roomID string) bool
	StartGame(roomID string, hostID string) bool
}

// roomService 实现 RoomService 接口
type roomService struct {
	roomRepo  repository.RoomRepository
	userRepo  repository.UserRepository
	resultRepo repository.ResultRepository
}

// NewRoomService 创建 RoomService 实例
func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository, resultRepo repository.ResultRepository) RoomService {
	return &roomService{
		roomRepo:  roomRepo,
		userRepo:  userRepo,
		resultRepo: resultRepo,
	}
}

// CreateRoom 处理创建房间逻辑
func (s *roomService) CreateRoom(req protocol.CreateRoomRequest, hostID string) (*models.Room, error) {
	// 创建新房间
	room := models.Room{
		ID:         "room_" + time.Now().Format("20060102150405"),
		Name:       req.Name,
		HostID:     hostID,
		Players:    []string{hostID},
		MaxPlayers: req.MaxPlayers,
		Status:     "waiting",
		CreatedAt:  time.Now(),
	}

	// 保存房间
	s.roomRepo.Add(room)

	// 更新用户的房间ID
	user := s.userRepo.FindByUsername(hostID)
	if user != nil {
		user.RoomID = room.ID
		s.userRepo.Update(hostID, *user)
	}

	return &room, nil
}

// JoinRoom 处理加入房间逻辑
func (s *roomService) JoinRoom(req protocol.JoinRoomRequest, username string) (*models.Room, string, error) {
	// 查找房间
	room := s.roomRepo.GetByID(req.RoomID)
	if room == nil {
		return nil, "房间不存在", nil
	}

	// 检查房间是否已满
	if len(room.Players) >= room.MaxPlayers {
		return nil, "房间已满", nil
	}

	// 检查房间状态
	if room.Status == "playing" {
		return nil, "游戏进行中，无法加入", nil
	}

	// 检查用户是否已在房间中
	for _, player := range room.Players {
		if player == username {
			return nil, "您已在房间中", nil
		}
	}

	// 添加用户到房间
	room.Players = append(room.Players, username)

	// 更新房间状态
	if len(room.Players) >= 2 {
		room.Status = "ready"
	}

	// 保存房间
	s.roomRepo.Update(*room)

	// 更新用户的房间ID
	user := s.userRepo.FindByUsername(username)
	if user != nil {
		user.RoomID = room.ID
		s.userRepo.Update(username, *user)
	}

	return room, "加入房间成功", nil
}

// GetRoomByID 根据ID获取房间
func (s *roomService) GetRoomByID(roomID string) *models.Room {
	return s.roomRepo.GetByID(roomID)
}

// GetAllRooms 获取所有房间
func (s *roomService) GetAllRooms() []models.Room {
	return s.roomRepo.GetAll()
}

// UpdateRoom 更新房间信息
func (s *roomService) UpdateRoom(room models.Room) bool {
	return s.roomRepo.Update(room)
}

// RemoveRoom 删除房间
func (s *roomService) RemoveRoom(roomID string) bool {
	return s.roomRepo.Remove(roomID)
}

// StartGame 处理开始游戏逻辑
func (s *roomService) StartGame(roomID string, hostID string) bool {
	// 查找房间
	room := s.roomRepo.GetByID(roomID)
	if room == nil {
		return false
	}

	// 检查是否是房主
	if room.HostID != hostID {
		return false
	}

	// 更新房间状态
	room.Status = "playing"
	s.roomRepo.Update(*room)

	return true
}