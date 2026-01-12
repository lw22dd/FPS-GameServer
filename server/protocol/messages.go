package protocol

import (
	"encoding/json"
)

// 这里是所有通信协议
type MessageType string

const (
	MsgTypeRegister       MessageType = "register"
	MsgTypeLogin          MessageType = "login"
	MsgTypeLoginResult    MessageType = "login_result"
	MsgTypeRegisterResult MessageType = "register_result"
	MsgTypeLogout         MessageType = "logout"
	MsgTypeHeartbeat      MessageType = "heartbeat"
	MsgTypeHeartbeatReply MessageType = "heartbeat_reply"
	MsgTypeCreateRoom     MessageType = "create_room"
	MsgTypeRoomList       MessageType = "room_list"
	MsgTypeJoinRoom       MessageType = "join_room"
	MsgTypeJoinRoomResult MessageType = "join_room_result"
	MsgTypeStartGame      MessageType = "start_game"
	MsgTypeGameStart      MessageType = "game_start"
	MsgTypeGameState      MessageType = "game_state"
	MsgTypePlayerAction   MessageType = "player_action"
	MsgTypeFire           MessageType = "fire"
	MsgTypeHit            MessageType = "hit"
	MsgTypeDeath          MessageType = "death"
	MsgTypeGameOver       MessageType = "game_over"
	MsgTypeError          MessageType = "error"
)

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type RoomInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Host       string   `json:"host"`
	Players    []string `json:"players"`
	MaxPlayers int      `json:"max_players"`
	Status     string   `json:"status"`
}

type RoomListResponse struct {
	Rooms []RoomInfo `json:"rooms"`
}

type CreateRoomRequest struct {
	Name       string `json:"name"`
	MaxPlayers int    `json:"max_players"`
}

type JoinRoomRequest struct {
	RoomID string `json:"room_id"`
}

type JoinRoomResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Room    RoomInfo `json:"room,omitempty"`
}

type CreateRoomResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	RoomID  string `json:"room_id"`
}

type PlayerAction struct {
	PlayerID string  `json:"player_id"`
	Action   string  `json:"action"`
	Value    float64 `json:"value"`
}

type FireAction struct {
	PlayerID  string  `json:"player_id"`
	Direction int     `json:"direction"`
	BulletID  string  `json:"bullet_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}

type HitAction struct {
	TargetID  string `json:"target_id"`
	Damage    int    `json:"damage"`
	Remaining int    `json:"remaining"`
}

type GameState struct {
	Hero1   HeroState     `json:"hero1"`
	Hero2   HeroState     `json:"hero2"`
	Bullets []BulletState `json:"bullets"`
	Status  string        `json:"status"`
}

type HeroState struct {
	ID        string  `json:"id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	HP        int     `json:"hp"`
	Direction int     `json:"direction"`
	Alive     bool    `json:"alive"`
}

type BulletState struct {
	ID      string  `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	VX      float64 `json:"vx"`
	OwnerID string  `json:"owner_id"`
}

type GameOverInfo struct {
	Winner   string `json:"winner"`
	Loser    string `json:"loser"`
	Duration int    `json:"duration"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
