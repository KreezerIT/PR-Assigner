package user

import "github.com/kreezerit/pr-assigner/internal/domain/user"

// SetIsActiveRequest HTTP‑запрос для изменения активности пользователя
type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// UserResponse HTTP‑ответ с данными пользователя
type UserResponse struct {
	User UserView `json:"user"`
}

// UserView представление пользователя в API
type UserView struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// ToDomainDTO преобразует HTTP‑запрос в DTO доменного слоя
func (r *SetIsActiveRequest) ToDomainDTO() *user.UpdateUserActivityDTO {
	return &user.UpdateUserActivityDTO{
		UserID:   r.UserID,
		IsActive: r.IsActive,
	}
}

// FromDomainDTO преобразует DTO доменного слоя в HTTP‑ответ
func FromDomainDTO(dto *user.DTO) *UserResponse {
	return &UserResponse{
		User: UserView{
			UserID:   dto.UserID,
			Username: dto.Username,
			TeamName: dto.TeamName,
			IsActive: dto.IsActive,
		},
	}
}
