package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/kreezerit/pr-assigner/internal/domain/statistics"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
)

// StatisticsRepository - операции для получения статистики
type StatisticsRepository struct {
	db *sql.DB
	qb sq.StatementBuilderType
}

func NewStatisticsRepository(db *sql.DB) *StatisticsRepository {
	return &StatisticsRepository{db: db, qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *StatisticsRepository) GetUserStats(ctx context.Context, userID string) (*statistics.UserStats, error) {
	// Метрики
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("stats_user").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Select(
		"u.user_id",
		"u.username",
		"t.team_name",
		"COUNT(prr.pull_request_id) as total_assignments",
		"COUNT(CASE WHEN s.name = 'OPEN' THEN 1 END) as active_reviews",
		"COUNT(CASE WHEN s.name = 'MERGED' THEN 1 END) as completed_reviews",
	).From("users u").
		Join("teams t ON u.team_id = t.id").
		LeftJoin("pull_request_reviewers prr ON u.user_id = prr.user_id").
		LeftJoin("pull_requests pr ON prr.pull_request_id = pr.pull_request_id").
		LeftJoin("pr_statuses s ON pr.status_id = s.id").
		Where(sq.Eq{"u.user_id": userID}).
		GroupBy("u.user_id", "u.username", "t.team_name")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var stats statistics.UserStats
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(
		&stats.UserID,
		&stats.Username,
		&stats.TeamName,
		&stats.TotalAssignments,
		&stats.ActiveReviews,
		&stats.CompletedReviews,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &statistics.UserStats{UserID: userID}, nil
		}
		return nil, fmt.Errorf("get user stats: %w", err)
	}
	return &stats, nil
}

func (r *StatisticsRepository) GetTeamStats(ctx context.Context, teamName string) (*statistics.TeamStats, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("stats_team").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Select(
		"t.team_name",
		"COUNT(DISTINCT u.user_id) as total_members",
		"COUNT(DISTINCT CASE WHEN u.is_active THEN u.user_id END) as active_members",
		"COUNT(DISTINCT pr.pull_request_id) as total_prs",
		"COUNT(DISTINCT CASE WHEN s.name = 'OPEN' THEN pr.pull_request_id END) as open_prs",
		"COUNT(DISTINCT CASE WHEN s.name = 'MERGED' THEN pr.pull_request_id END) as merged_prs",
	).From("teams t").
		LeftJoin("users u ON t.id = u.team_id").
		LeftJoin("pull_requests pr ON u.user_id = pr.author_id").
		LeftJoin("pr_statuses s ON pr.status_id = s.id").
		Where(sq.Eq{"t.team_name": teamName}).
		GroupBy("t.team_name")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var stats statistics.TeamStats
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(
		&stats.TeamName,
		&stats.TotalMembers,
		&stats.ActiveMembers,
		&stats.TotalPRs,
		&stats.OpenPRs,
		&stats.MergedPRs,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &statistics.TeamStats{TeamName: teamName}, nil
		}
		return nil, fmt.Errorf("get team stats: %w", err)
	}
	return &stats, nil
}

func (r *StatisticsRepository) GetGlobalStats(ctx context.Context) (*statistics.GlobalStats, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("stats_global").Observe(time.Since(start).Seconds()) }()

	var stats statistics.GlobalStats

	userQuery := r.qb.Select("COUNT(*) as total", "COUNT(CASE WHEN is_active THEN 1 END) as active").
		From("users")
	sqlStr, args, err := userQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build user query: %w", err)
	}
	if err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&stats.TotalUsers, &stats.ActiveUsers); err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	teamQuery := r.qb.Select("COUNT(*)").
		From("teams")
	sqlStr, args, err = teamQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build team query: %w", err)
	}
	if err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&stats.TotalTeams); err != nil {
		return nil, fmt.Errorf("get team stats: %w", err)
	}

	prQuery := r.qb.Select(
		"COUNT(*) as total",
		"COUNT(CASE WHEN s.name = 'OPEN' THEN 1 END) as open",
		"COUNT(CASE WHEN s.name = 'MERGED' THEN 1 END) as merged",
	).From("pull_requests pr").
		Join("pr_statuses s ON pr.status_id = s.id")
	sqlStr, args, err = prQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build pr query: %w", err)
	}
	if err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&stats.TotalPRs, &stats.OpenPRs, &stats.MergedPRs); err != nil {
		return nil, fmt.Errorf("get pr stats: %w", err)
	}

	assignmentsQuery := r.qb.Select("COUNT(*)").
		From("pull_request_reviewers")
	sqlStr, args, err = assignmentsQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build assignments query: %w", err)
	}
	if err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&stats.TotalAssignments); err != nil {
		return nil, fmt.Errorf("get assignments stats: %w", err)
	}

	return &stats, nil
}

func (r *StatisticsRepository) GetTopReviewers(ctx context.Context, limit int) ([]statistics.TopReviewer, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("stats_top_reviewers").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Select(
		"u.user_id",
		"u.username",
		"t.team_name",
		"COUNT(prr.pull_request_id) as review_count",
	).From("users u").
		Join("teams t ON u.team_id = t.id").
		LeftJoin("pull_request_reviewers prr ON u.user_id = prr.user_id").
		GroupBy("u.user_id", "u.username", "t.team_name").
		Having("COUNT(prr.pull_request_id) > 0").
		OrderBy("review_count DESC")

	if limit > 0 {
		query = query.Limit(uint64(limit))
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("get top reviewers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var reviewers []statistics.TopReviewer
	for rows.Next() {
		var reviewer statistics.TopReviewer
		if err := rows.Scan(&reviewer.UserID, &reviewer.Username, &reviewer.TeamName, &reviewer.ReviewCount); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviewers: %w", err)
	}
	return reviewers, nil
}
