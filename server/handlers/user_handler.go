package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"game/data"
	"game/models"
	"game/protocol"
)

type UserHandler struct {
	userStore *data.UserStore
}

func NewUserHandler(userStore *data.UserStore) *UserHandler {
	return &UserHandler{userStore: userStore}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req protocol.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if strings.Contains(req.Username, " ") {
		h.sendRegisterResponse(w, false, "用户名不能包含空格")
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 20 {
		h.sendRegisterResponse(w, false, "用户名长度必须在3-20之间")
		return
	}

	if len(req.Password) < 6 {
		h.sendRegisterResponse(w, false, "密码长度至少6位")
		return
	}

	if !strings.Contains(req.Email, "@") {
		h.sendRegisterResponse(w, false, "邮箱格式不正确")
		return
	}

	existingUser := h.userStore.FindByUsername(req.Username)
	if existingUser != nil {
		h.sendRegisterResponse(w, false, "用户名已存在")
		return
	}

	existingEmail := h.userStore.FindByEmail(req.Email)
	if existingEmail != nil {
		h.sendRegisterResponse(w, false, "该邮箱已被注册")
		return
	}

	user := models.User{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Online:    false,
		LoginTime: time.Time{},
		RoomID:    "",
	}

	h.userStore.Add(user)
	h.sendRegisterResponse(w, true, "注册成功")
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req protocol.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	user := h.userStore.FindByUsername(req.Username)
	if user == nil {
		h.sendLoginResponse(w, false, "用户不存在", "")
		return
	}

	if user.Password != req.Password {
		h.sendLoginResponse(w, false, "密码错误", "")
		return
	}

	if user.Online {
		h.sendLoginResponse(w, false, "用户已登录", "")
		return
	}

	user.Online = true
	user.LoginTime = time.Now()
	user.RoomID = ""
	h.userStore.Update(req.Username, *user)

	h.sendLoginResponse(w, true, "登录成功", req.Username)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	user := h.userStore.FindByUsername(req.Username)
	if user != nil {
		user.Online = false
		user.RoomID = ""
		h.userStore.Update(req.Username, *user)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "已退出登录"})
}

func (h *UserHandler) sendRegisterResponse(w http.ResponseWriter, success bool, message string) {
	w.Header().Set("Content-Type", "application/json")
	resp := protocol.RegisterResponse{
		Success: success,
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) sendLoginResponse(w http.ResponseWriter, success bool, message, token string) {
	w.Header().Set("Content-Type", "application/json")
	resp := protocol.LoginResponse{
		Success: success,
		Message: message,
		Token:   token,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := protocol.ErrorResponse{
		Code:    status,
		Message: message,
	}
	json.NewEncoder(w).Encode(resp)
}

func RegisterRoutes(mux *http.ServeMux, userStore *data.UserStore) {
	handler := NewUserHandler(userStore)
	mux.HandleFunc("/user/register", handler.Register)
	mux.HandleFunc("/user/login", handler.Login)
	mux.HandleFunc("/user/logout", handler.Logout)
}
