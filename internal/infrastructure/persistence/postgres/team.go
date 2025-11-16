package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
)

// TeamRepository - операции над таблицей команд в БД
type TeamRepository struct {
	db *sql.DB
	qb sq.StatementBuilderType
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{
		db: db,
		qb: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TeamRepository) Create(ctx context.Context, t *team.Team) error {
	// Метрики
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("team_create").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Insert("teams").
		Columns("team_name").
		Values(t.TeamName).
		Suffix("RETURNING id, created_at")

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, teamName string) (*team.Team, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueryDuration.WithLabelValues("team_get_by_name").Observe(time.Since(start).Seconds())
	}()

	// Получаем команду
	teamQuery := r.qb.Select("id", "team_name", "created_at").
		From("teams").
		Where(sq.Eq{"team_name": teamName})

	sqlStr, args, err := teamQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build team query: %w", err)
	}

	var t team.Team
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&t.ID, &t.TeamName, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get team: %w", err)
	}

	// Получаем участников команды
	membersQuery := r.qb.Select("user_id", "username", "is_active").
		From("users").
		Where(sq.Eq{"team_id": t.ID}).
		OrderBy("username")

	sqlStr, args, err = membersQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build members query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("get members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	t.Members = []team.Member{}
	for rows.Next() {
		var m team.Member
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		t.Members = append(t.Members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return &t, nil
}

func (r *TeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("team_exists").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Select("1").
		From("teams").
		Where(sq.Eq{"team_name": teamName}).
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

func (r *TeamRepository) GetIDByName(ctx context.Context, teamName string) (int, error) {
	start := time.Now()
	defer func() { metrics.DBQueryDuration.WithLabelValues("team_get_id").Observe(time.Since(start).Seconds()) }()

	query := r.qb.Select("id").
		From("teams").
		Where(sq.Eq{"team_name": teamName})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var id int
	err = r.db.QueryRowContext(ctx, sqlStr, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("get team id: %w", err)
	}

	return id, nil
}
