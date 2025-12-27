package models

import "time"

type User struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Email      string    `json:"email"`
	Online     bool      `json:"online"`
	LoginTime  time.Time `json:"login_time"`
	RoomID     string    `json:"room_id"`
}

type UsersData struct {
	Users []User `json:"users"`
}

type Room struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	HostID      string        `json:"host_id"`
	Players     []string      `json:"players"`
	MaxPlayers  int           `json:"max_players"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
}

type RoomsData struct {
	Rooms []Room `json:"rooms"`
}

type GameResult struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	Winner      string    `json:"winner"`
	Loser       string    `json:"loser"`
	PlayTime    time.Time `json:"play_time"`
	Duration    int       `json:"duration"`
}

type GameResultsData struct {
	Results []GameResult `json:"results"`
}

type HeroState struct {
	ID       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	HP       int     `json:"hp"`
	Direction int    `json:"direction"`
	Alive    bool    `json:"alive"`
}

type GameState struct {
	RoomID     string      `json:"room_id"`
	Hero1      HeroState   `json:"hero1"`
	Hero2      HeroState   `json:"hero2"`
	Bullets    []Bullet    `json:"bullets"`
	GameStatus string      `json:"game_status"`
}

type Bullet struct {
	ID       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	VX       float64 `json:"vx"`
	OwnerID  string  `json:"owner_id"`
}
