package user

import "context"

type Repository interface {
	Upsert(ctx context.Context, user *User) error
	GetByID(ctx context.Context, userID string) (*User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserIDs []string) ([]User, error)
}
