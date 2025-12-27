package repository

import (
	"game/data"
	"game/models"
)

// ResultRepository 定义游戏结果数据访问接口
type ResultRepository interface {
	Add(result models.GameResult)
	GetAll() []models.GameResult
}

// resultRepository 实现 ResultRepository 接口
type resultRepository struct {
	store *data.ResultStore
}

// NewResultRepository 创建 ResultRepository 实例
func NewResultRepository(store *data.ResultStore) ResultRepository {
	return &resultRepository{store: store}
}

// Add 添加游戏结果
func (r *resultRepository) Add(result models.GameResult) {
	r.store.Add(result)
}

// GetAll 获取所有游戏结果
func (r *resultRepository) GetAll() []models.GameResult {
	return r.store.GetAll()
}