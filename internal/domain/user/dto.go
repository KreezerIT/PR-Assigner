package user

// DTO информация о пользователе
type DTO struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}

// UpdateUserActivityDTO используется для обновления активности пользователя
type UpdateUserActivityDTO struct {
	UserID   string
	IsActive bool
}

// ToDTO преобразует доменную модель User в DTO
func (u *User) ToDTO() *DTO {
	return &DTO{
		UserID:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}
