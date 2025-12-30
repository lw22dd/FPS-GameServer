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

func setupUserTestStore(t *testing.T) *data.UserStore {
	store := data.NewUserStore()
	store.GetAll()
	for _, user := range store.GetAll() {
		store.Update(user.Username, user)
	}
	return store
}

func TestRegister_Success(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("Expected success true, got false: %s", resp.Message)
	}
	if resp.Message != "注册成功" {
		t.Errorf("Expected message '注册成功', got '%s'", resp.Message)
	}

	user := store.FindByUsername("testuser")
	if user == nil {
		t.Fatal("Expected user to be created, but not found")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestRegister_UsernameWithSpace(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.RegisterRequest{
		Username: "test user",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "用户名不能包含空格" {
		t.Errorf("Expected message '用户名不能包含空格', got '%s'", resp.Message)
	}
}

func TestRegister_UsernameTooShort(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.RegisterRequest{
		Username: "ab",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "用户名长度必须在3-20之间" {
		t.Errorf("Expected message '用户名长度必须在3-20之间', got '%s'", resp.Message)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.RegisterRequest{
		Username: "testuser",
		Password: "12345",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "密码长度至少6位" {
		t.Errorf("Expected message '密码长度至少6位', got '%s'", resp.Message)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "invalid-email",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "邮箱格式不正确" {
		t.Errorf("Expected message '邮箱格式不正确', got '%s'", resp.Message)
	}
}

func TestRegister_UsernameExists(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	existingUser := models.User{
		Username: "existing",
		Password: "password123",
		Email:    "existing@example.com",
	}
	store.Add(existingUser)

	reqBody := protocol.RegisterRequest{
		Username: "existing",
		Password: "password123",
		Email:    "new@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "用户名已存在" {
		t.Errorf("Expected message '用户名已存在', got '%s'", resp.Message)
	}
}

func TestRegister_EmailExists(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	existingUser := models.User{
		Username: "existing",
		Password: "password123",
		Email:    "existing@example.com",
	}
	store.Add(existingUser)

	reqBody := protocol.RegisterRequest{
		Username: "newuser",
		Password: "password123",
		Email:    "existing@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	var resp protocol.RegisterResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "该邮箱已被注册" {
		t.Errorf("Expected message '该邮箱已被注册', got '%s'", resp.Message)
	}
}

func TestLogin_Success(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	user := models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	store.Add(user)

	reqBody := protocol.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp protocol.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("Expected success true, got false: %s", resp.Message)
	}
	if resp.Message != "登录成功" {
		t.Errorf("Expected message '登录成功', got '%s'", resp.Message)
	}
	if resp.Token != "testuser" {
		t.Errorf("Expected token 'testuser', got '%s'", resp.Token)
	}

	loggedInUser := store.FindByUsername("testuser")
	if loggedInUser == nil {
		t.Fatal("Expected user to be found")
	}
	if !loggedInUser.Online {
		t.Error("Expected user to be online")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := protocol.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	var resp protocol.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "用户不存在" {
		t.Errorf("Expected message '用户不存在', got '%s'", resp.Message)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	user := models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	store.Add(user)

	reqBody := protocol.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	var resp protocol.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "密码错误" {
		t.Errorf("Expected message '密码错误', got '%s'", resp.Message)
	}
}

func TestLogin_AlreadyLoggedIn(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	user := models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Online:   true,
	}
	store.Add(user)

	reqBody := protocol.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	var resp protocol.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Success {
		t.Error("Expected success false, got true")
	}
	if resp.Message != "用户已登录" {
		t.Errorf("Expected message '用户已登录', got '%s'", resp.Message)
	}
}

func TestLogout_Success(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	user := models.User{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
		Online:   true,
	}
	store.Add(user)

	reqBody := struct {
		Username string `json:"username"`
	}{
		Username: "testuser",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	loggedOutUser := store.FindByUsername("testuser")
	if loggedOutUser == nil {
		t.Fatal("Expected user to be found")
	}
	if loggedOutUser.Online {
		t.Error("Expected user to be offline")
	}
}

func TestLogout_UserNotFound(t *testing.T) {
	store := setupUserTestStore(t)
	handler := handlers.NewUserHandler(store)

	reqBody := struct {
		Username string `json:"username"`
	}{
		Username: "nonexistent",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/user/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}
