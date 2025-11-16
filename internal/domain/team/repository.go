package team

import "context"

type Repository interface {
	Create(ctx context.Context, team *Team) error
	GetByName(ctx context.Context, teamName string) (*Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
	GetIDByName(ctx context.Context, teamName string) (int, error)
}
