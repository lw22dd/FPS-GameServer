package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"game/models"
)

var (
	DataDir string
)

func init() {
	// 直接使用当前目录下的 data 目录，而不是基于可执行文件的位置
	DataDir = "data"
	fmt.Printf("数据目录: %s\n", DataDir)
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		fmt.Printf("创建数据目录失败: %v\n", err)
	}
}

type UserStore struct {
	mu    sync.RWMutex
	users []models.User
	file  string
}

type RoomStore struct {
	mu    sync.RWMutex
	rooms []models.Room
	file  string
}

type ResultStore struct {
	mu      sync.RWMutex
	results []models.GameResult
	file    string
}

func NewUserStore() *UserStore {
	file := filepath.Join(DataDir, "users.json")
	store := &UserStore{
		users: make([]models.User, 0),
		file:  file,
	}
	store.load()
	return store
}

func NewRoomStore() *RoomStore {
	file := filepath.Join(DataDir, "rooms.json")
	store := &RoomStore{
		rooms: make([]models.Room, 0),
		file:  file,
	}
	store.load()
	return store
}

func NewResultStore() *ResultStore {
	file := filepath.Join(DataDir, "game_results.json")
	store := &ResultStore{
		results: make([]models.GameResult, 0),
		file:    file,
	}
	store.load()
	return store
}

func (s *UserStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			s.users = make([]models.User, 0)
			return
		}
		fmt.Printf("加载用户数据失败: %v\n", err)
		return
	}
	var usersData models.UsersData
	if err := json.Unmarshal(data, &usersData); err != nil {
		fmt.Printf("解析用户数据失败: %v\n", err)
		s.users = make([]models.User, 0)
		return
	}
	s.users = usersData.Users
}

func (s *UserStore) save() {
	usersData := models.UsersData{Users: s.users}
	data, err := json.MarshalIndent(usersData, "", "  ")
	if err != nil {
		fmt.Printf("序列化用户数据失败: %v\n", err)
		return
	}
	if err := os.WriteFile(s.file, data, 0644); err != nil {
		fmt.Printf("保存用户数据失败: %v\n", err)
	}
}

func (s *UserStore) Add(user models.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = append(s.users, user)
	s.save()
}

func (s *UserStore) FindByUsername(username string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.users {
		if s.users[i].Username == username {
			return &s.users[i]
		}
	}
	return nil
}

func (s *UserStore) FindByEmail(email string) *models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.users {
		if s.users[i].Email == email {
			return &s.users[i]
		}
	}
	return nil
}

func (s *UserStore) Update(username string, user models.User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.users {
		if s.users[i].Username == username {
			s.users[i] = user
			s.save()
			return true
		}
	}
	return false
}

func (s *UserStore) GetAll() []models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.User, len(s.users))
	copy(result, s.users)
	return result
}

func (s *RoomStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			s.rooms = make([]models.Room, 0)
			return
		}
		fmt.Printf("加载房间数据失败: %v\n", err)
		return
	}
	var roomsData models.RoomsData
	if err := json.Unmarshal(data, &roomsData); err != nil {
		fmt.Printf("解析房间数据失败: %v\n", err)
		s.rooms = make([]models.Room, 0)
		return
	}
	s.rooms = roomsData.Rooms
}

func (s *RoomStore) save() {
	roomsData := models.RoomsData{Rooms: s.rooms}
	data, err := json.MarshalIndent(roomsData, "", "  ")
	if err != nil {
		fmt.Printf("序列化房间数据失败: %v\n", err)
		return
	}
	if err := os.WriteFile(s.file, data, 0644); err != nil {
		fmt.Printf("保存房间数据失败: %v\n", err)
	}
}

func (s *RoomStore) Add(room models.Room) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms = append(s.rooms, room)
	s.save()
}

func (s *RoomStore) GetByID(id string) *models.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.rooms {
		if s.rooms[i].ID == id {
			return &s.rooms[i]
		}
	}
	return nil
}

func (s *RoomStore) GetAll() []models.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.Room, len(s.rooms))
	copy(result, s.rooms)
	return result
}

func (s *RoomStore) Update(room models.Room) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rooms {
		if s.rooms[i].ID == room.ID {
			s.rooms[i] = room
			s.save()
			return true
		}
	}
	return false
}

func (s *RoomStore) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rooms {
		if s.rooms[i].ID == id {
			s.rooms = append(s.rooms[:i], s.rooms[i+1:]...)
			s.save()
			return true
		}
	}
	return false
}

func (s *ResultStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			s.results = make([]models.GameResult, 0)
			return
		}
		fmt.Printf("加载游戏结果数据失败: %v\n", err)
		return
	}
	var resultsData models.GameResultsData
	if err := json.Unmarshal(data, &resultsData); err != nil {
		fmt.Printf("解析游戏结果数据失败: %v\n", err)
		s.results = make([]models.GameResult, 0)
		return
	}
	s.results = resultsData.Results
}

func (s *ResultStore) save() {
	resultsData := models.GameResultsData{Results: s.results}
	data, err := json.MarshalIndent(resultsData, "", "  ")
	if err != nil {
		fmt.Printf("序列化游戏结果数据失败: %v\n", err)
		return
	}
	if err := os.WriteFile(s.file, data, 0644); err != nil {
		fmt.Printf("保存游戏结果数据失败: %v\n", err)
	}
}

func (s *ResultStore) Add(result models.GameResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = append(s.results, result)
	s.save()
}

func (s *ResultStore) GetAll() []models.GameResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.GameResult, len(s.results))
	copy(result, s.results)
	return result
}
