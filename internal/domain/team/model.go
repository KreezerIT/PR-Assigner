package team

import "time"

// Team представляет команду
type Team struct {
	ID        int       `json:"-"`
	TeamName  string    `json:"team_name"`
	Members   []Member  `json:"members"`
	CreatedAt time.Time `json:"-"`
}

// Member представляет участника команды
type Member struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
