package user

import "time"

// User представляет пользователя
type User struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	TeamID    int       `json:"-"`
	TeamName  string    `json:"team_name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
