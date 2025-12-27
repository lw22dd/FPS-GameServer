package api

import (
	"game/protocol"
	"game/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler 定义用户 API 处理函数结构
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建 UserHandler 实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Register 处理用户注册请求
func (h *UserHandler) Register(c *gin.Context) {
	var req protocol.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 调用 Service 层处理注册逻辑
	success, message := h.userService.Register(req)

	// 返回响应
	c.JSON(http.StatusOK, protocol.RegisterResponse{
		Success: success,
		Message: message,
	})
}

// Login 处理用户登录请求
func (h *UserHandler) Login(c *gin.Context) {
	var req protocol.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 调用 Service 层处理登录逻辑
	success, message, token := h.userService.Login(req)

	// 返回响应
	c.JSON(http.StatusOK, protocol.LoginResponse{
		Success: success,
		Message: message,
		Token:   token,
	})
}

// Logout 处理用户登出请求
func (h *UserHandler) Logout(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, protocol.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "请求格式错误",
		})
		return
	}

	// 调用 Service 层处理登出逻辑
	h.userService.Logout(req.Username)

	// 返回响应
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}

// Test 处理测试请求
func (h *UserHandler) Test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "服务器运行正常"})
}