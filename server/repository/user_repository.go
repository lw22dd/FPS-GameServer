package repository

import (
	"game/data"
	"game/models"
)

// UserRepository 定义用户数据访问接口
type UserRepository interface {
	Add(user models.User)
	FindByUsername(username string) *models.User
	FindByEmail(email string) *models.User
	Update(username string, user models.User) bool
	GetAll() []models.User
}

// userRepository 实现 UserRepository 接口
type userRepository struct {
	store *data.UserStore
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(store *data.UserStore) UserRepository {
	return &userRepository{store: store}
}

// Add 添加用户
func (r *userRepository) Add(user models.User) {
	r.store.Add(user)
}

// FindByUsername 根据用户名查找用户
func (r *userRepository) FindByUsername(username string) *models.User {
	return r.store.FindByUsername(username)
}

// FindByEmail 根据邮箱查找用户
func (r *userRepository) FindByEmail(email string) *models.User {
	return r.store.FindByEmail(email)
}

// Update 更新用户信息
func (r *userRepository) Update(username string, user models.User) bool {
	return r.store.Update(username, user)
}

// GetAll 获取所有用户
func (r *userRepository) GetAll() []models.User {
	return r.store.GetAll()
}