package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
)

// PullRequestRepository - операции над таблицей ПР в БД
type PullRequestRepository struct {
	db *sql.DB
	qb sq.StatementBuilderType
}

func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PullRequestRepository) Create(ctx context.Context, pr *pullrequest.PullRequest) error {
	// Метрики
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("pr_create").Observe(time.Since(start).Seconds()) }()

	// Начало транзакции (для вставки PR и списка ревьюверов)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Получаем ID статуса `OPEN`
	var statusID int
	err = tx.QueryRowContext(ctx, "SELECT id FROM pr_statuses WHERE name = $1", pullrequest.StatusOpen).Scan(&statusID)
	if err != nil {
		return fmt.Errorf("get status id: %w", err)
	}

	// Вставка PR
	query := r.qb.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status_id").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, statusID)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build pr query: %w", err)
	}

	_, err = tx.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("insert pr: %w", err)
	}

	// Вставка ревьюверов
	if len(pr.AssignedReviewers) > 0 {
		reviewersQuery := r.qb.Insert("pull_request_reviewers").Columns("pull_request_id", "user_id")
		for _, reviewerID := range pr.AssignedReviewers {
			reviewersQuery = reviewersQuery.Values(pr.PullRequestID, reviewerID)
		}
		sqlStr, args, err = reviewersQuery.ToSql()
		if err != nil {
			return fmt.Errorf("build reviewers query: %w", err)
		}
		_, err = tx.ExecContext(ctx, sqlStr, args...)
		if err != nil {
			return fmt.Errorf("insert reviewers: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) GetByID(ctx context.Context, prID string) (*pullrequest.PullRequest, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("pr_get_by_id").Observe(time.Since(start).Seconds()) }()

	// Получаем PR
	query := r.qb.Select(
		"pr.pull_request_id",
		"pr.pull_request_name",
		"pr.author_id",
		"s.name",
		"pr.created_at",
		"pr.merged_at",
	).From("pull_requests pr").
		Join("pr_statuses s ON pr.status_id = s.id").
		Where(sq.Eq{"pr.pull_request_id": prID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var pr pullrequest.PullRequest
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get pr: %w", err)
	}

	// Получаем ревьюверов
	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PullRequestRepository) Exists(ctx context.Context, prID string) (bool, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("pr_exists").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Select("1").
		From("pull_requests").
		Where(sq.Eq{"pull_request_id": prID}).
		Limit(1)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("build query: %w", err)
	}

	var exists int
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check existence: %w", err)
	}
	return true, nil
}

func (r *PullRequestRepository) Merge(ctx context.Context, prID string) error {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("pr_merge").Observe(time.Since(start).Seconds()) }()

	// Получаем ID статуса `MERGED`
	var statusID int
	err := r.db.QueryRowContext(ctx, "SELECT id FROM pr_statuses WHERE name = $1", pullrequest.StatusMerged).Scan(&statusID)
	if err != nil {
		return fmt.Errorf("get merged status id: %w", err)
	}

	query := r.qb.Update("pull_requests").Set("status_id", statusID).Set("merged_at", time.Now()).Where(sq.Eq{"pull_request_id": prID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("merge pr: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	query := r.qb.Select("user_id").
		From("pull_request_reviewers").
		Where(sq.Eq{"pull_request_id": prID}).
		OrderBy("assigned_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	reviewers := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}
	return reviewers, nil
}

func (r *PullRequestRepository) RemoveReviewer(ctx context.Context, prID string, userID string) error {
	query := r.qb.Delete("pull_request_reviewers").Where(sq.Eq{"pull_request_id": prID, "user_id": userID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("remove reviewer: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) AddReviewer(ctx context.Context, prID string, userID string) error {
	query := r.qb.Insert("pull_request_reviewers").Columns("pull_request_id", "user_id").Values(prID, userID)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("add reviewer: %w", err)
	}
	return nil
}

func (r *PullRequestRepository) GetByReviewer(ctx context.Context, userID string) ([]pullrequest.PullRequestShort, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("pr_get_by_reviewer").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Select(
		"pr.pull_request_id",
		"pr.pull_request_name",
		"pr.author_id",
		"s.name",
	).From("pull_requests pr").
		Join("pr_statuses s ON pr.status_id = s.id").
		Join("pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where(sq.Eq{"prr.user_id": userID}).
		OrderBy("pr.created_at DESC")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query prs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var prs []pullrequest.PullRequestShort
	for rows.Next() {
		var pr pullrequest.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prs: %w", err)
	}
	return prs, nil
}

func (r *PullRequestRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("pr_replace_reviewer").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Update("pull_request_reviewers").Set("user_id", newUserID).Set("assigned_at", time.Now()).Where(sq.Eq{"pull_request_id": prID, "user_id": oldUserID})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("replace reviewer: %w", err)
	}
	return nil
}
