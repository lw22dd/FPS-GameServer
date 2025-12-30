package repository

import (
	"game/data"
	"game/models"
)

// RoomRepository 定义房间数据访问接口
type RoomRepository interface {
	Add(room models.Room)
	GetByID(id string) *models.Room
	GetAll() []models.Room
	Update(room models.Room) bool
	Remove(id string) bool
}

// roomRepository 实现 RoomRepository 接口
type roomRepository struct {
	store *data.RoomStore
}

// NewRoomRepository 创建 RoomRepository 实例
func NewRoomRepository(store *data.RoomStore) RoomRepository {
	return &roomRepository{store: store}
}

// Add 添加房间
func (r *roomRepository) Add(room models.Room) {
	r.store.Add(room)
}

// GetByID 根据ID查找房间
func (r *roomRepository) GetByID(id string) *models.Room {
	return r.store.GetByID(id)
}

// GetAll 获取所有房间
func (r *roomRepository) GetAll() []models.Room {
	return r.store.GetAll()
}

// Update 更新房间信息
func (r *roomRepository) Update(room models.Room) bool {
	return r.store.Update(room)
}

// Remove 删除房间
func (r *roomRepository) Remove(id string) bool {
	return r.store.Remove(id)
}