package service

import (
	"game/models"
	"game/protocol"
	"game/repository"
	"strings"
	"time"
)

// UserService 定义用户业务逻辑接口
type UserService interface {
	Register(req protocol.RegisterRequest) (bool, string)
	Login(req protocol.LoginRequest) (bool, string, string)
	Logout(username string)
}

// userService 实现 UserService 接口
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// Register 处理用户注册逻辑
func (s *userService) Register(req protocol.RegisterRequest) (bool, string) {
	// 验证用户名是否包含空格
	if strings.Contains(req.Username, " ") {
		return false, "用户名不能包含空格"
	}

	// 验证用户名长度
	if len(req.Username) < 3 || len(req.Username) > 20 {
		return false, "用户名长度必须在3-20之间"
	}

	// 验证密码长度
	if len(req.Password) < 6 {
		return false, "密码长度至少6位"
	}

	// 验证邮箱格式
	if !strings.Contains(req.Email, "@") {
		return false, "邮箱格式不正确"
	}

	// 检查用户名是否已存在
	if existingUser := s.userRepo.FindByUsername(req.Username); existingUser != nil {
		return false, "用户名已存在"
	}

	// 检查邮箱是否已被注册
	if existingEmail := s.userRepo.FindByEmail(req.Email); existingEmail != nil {
		return false, "该邮箱已被注册"
	}

	// 创建新用户
	user := models.User{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Online:    false,
		LoginTime: time.Time{},
		RoomID:    "",
	}

	// 保存用户
	s.userRepo.Add(user)
	return true, "注册成功"
}

// Login 处理用户登录逻辑
func (s *userService) Login(req protocol.LoginRequest) (bool, string, string) {
	// 查找用户
	user := s.userRepo.FindByUsername(req.Username)
	if user == nil {
		return false, "用户不存在", ""
	}

	// 验证密码
	if user.Password != req.Password {
		return false, "密码错误", ""
	}

	// 检查用户是否已在线
	if user.Online {
		return false, "用户已登录", ""
	}

	// 更新用户状态
	user.Online = true
	user.LoginTime = time.Now()
	user.RoomID = ""
	s.userRepo.Update(req.Username, *user)

	return true, "登录成功", req.Username
}

// Logout 处理用户登出逻辑
func (s *userService) Logout(username string) {
	user := s.userRepo.FindByUsername(username)
	if user != nil {
		user.Online = false
		user.RoomID = ""
		s.userRepo.Update(username, *user)
	}
}