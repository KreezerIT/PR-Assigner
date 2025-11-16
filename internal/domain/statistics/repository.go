package statistics

import "context"

type Repository interface {
	GetUserStats(ctx context.Context, userID string) (*UserStats, error)
	GetTeamStats(ctx context.Context, teamName string) (*TeamStats, error)
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetTopReviewers(ctx context.Context, limit int) ([]TopReviewer, error)
}
