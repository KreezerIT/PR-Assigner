package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
)

// UserRepository - операции над таблицей пользователей в БД
type UserRepository struct {
	db *sql.DB
	qb sq.StatementBuilderType
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) Upsert(ctx context.Context, u *user.User) error {
	// Метрики
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("user_upsert").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Insert("users").
		Columns("user_id", "username", "team_id", "is_active", "updated_at").
		Values(u.UserID, u.Username, u.TeamID, u.IsActive, time.Now()).
		Suffix(`
            ON CONFLICT (user_id) 
            DO UPDATE SET 
                username = EXCLUDED.username,
                team_id = EXCLUDED.team_id,
                is_active = EXCLUDED.is_active,
                updated_at = EXCLUDED.updated_at
        `)

	toSQL, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, toSQL, args...)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*user.User, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("user_get_by_id").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Select(
		"u.user_id",
		"u.username",
		"u.team_id",
		"t.team_name",
		"u.is_active",
		"u.created_at",
		"u.updated_at",
	).
		From("users u").
		Join("teams t ON u.team_id = t.id").
		Where(sq.Eq{"u.user_id": userID})

	toSQL, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var u user.User
	err = r.db.QueryRowContext(ctx, toSQL, args...).Scan(
		&u.UserID,
		&u.Username,
		&u.TeamID,
		&u.TeamName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("user_set_active").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Update("users").
		Set("is_active", isActive).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"user_id": userID})

	toSQL, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	_, err = r.db.ExecContext(ctx, toSQL, args...)
	if err != nil {
		return fmt.Errorf("update is_active: %w", err)
	}

	return nil
}

func (r *UserRepository) GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserIDs []string) ([]user.User, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("user_get_active_team").Observe(time.Since(start).Seconds())
	}()

	query := r.qb.Select(
		"u.user_id",
		"u.username",
		"u.team_id",
		"t.team_name",
		"u.is_active",
		"u.created_at",
		"u.updated_at",
	).
		From("users u").
		Join("teams t ON u.team_id = t.id").
		Where(sq.Eq{"u.team_id": teamID, "u.is_active": true})

	if len(excludeUserIDs) > 0 {
		query = query.Where(sq.NotEq{"u.user_id": excludeUserIDs})
	}

	toSQL, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, toSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(
			&u.UserID,
			&u.Username,
			&u.TeamID,
			&u.TeamName,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}
