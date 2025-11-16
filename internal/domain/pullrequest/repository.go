package pullrequest

import "context"

type Repository interface {
	Create(ctx context.Context, pr *PullRequest) error
	GetByID(ctx context.Context, prID string) (*PullRequest, error)
	Exists(ctx context.Context, prID string) (bool, error)
	Merge(ctx context.Context, prID string) error
	GetReviewers(ctx context.Context, prID string) ([]string, error)
	RemoveReviewer(ctx context.Context, prID string, userID string) error
	AddReviewer(ctx context.Context, prID string, userID string) error
	GetByReviewer(ctx context.Context, userID string) ([]PullRequestShort, error)
	ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
}
