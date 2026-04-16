package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRow struct {
	ID        int64
	UserID    pgtype.Int8
	Name      string
	Type      string
	Color     pgtype.Text
	Icon      pgtype.Text
	IsSystem  bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) ListForUser(ctx context.Context, userID int64) ([]CategoryRow, error) {
	const q = `
SELECT id, user_id, name, type, color, icon, is_system, created_at, updated_at
FROM categories
WHERE is_system = TRUE OR user_id = $1
ORDER BY is_system DESC, name ASC
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CategoryRow, 0)
	for rows.Next() {
		var c CategoryRow
		if err = rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.Color, &c.Icon, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoryRepository) Create(ctx context.Context, userID int64, name, catType string, color, icon *string) (CategoryRow, error) {
	const q = `
INSERT INTO categories (user_id, name, type, color, icon)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, type, color, icon, is_system, created_at, updated_at
`
	var c CategoryRow
	err := r.pool.QueryRow(ctx, q, userID, name, catType, textFromPtr(color), textFromPtr(icon)).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.Color, &c.Icon, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *CategoryRepository) GetByIDForUser(ctx context.Context, id, userID int64) (CategoryRow, error) {
	const q = `
SELECT id, user_id, name, type, color, icon, is_system, created_at, updated_at
FROM categories
WHERE id = $1 AND (is_system = TRUE OR user_id = $2)
`
	var c CategoryRow
	err := r.pool.QueryRow(ctx, q, id, userID).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.Color, &c.Icon, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *CategoryRepository) Update(ctx context.Context, id, userID int64, name, color, icon *string) (CategoryRow, error) {
	const q = `
UPDATE categories
SET name      = COALESCE($3::text, name),
    color     = COALESCE($4::text, color),
    icon      = COALESCE($5::text, icon),
    updated_at = $6
WHERE id = $1 AND user_id = $2 AND is_system = FALSE
RETURNING id, user_id, name, type, color, icon, is_system, created_at, updated_at
`
	var c CategoryRow
	err := r.pool.QueryRow(ctx, q, id, userID, textFromPtr(name), textFromPtr(color), textFromPtr(icon), time.Now().UTC()).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Type, &c.Color, &c.Icon, &c.IsSystem, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *CategoryRepository) Delete(ctx context.Context, id, userID int64) (int64, error) {
	const q = `DELETE FROM categories WHERE id = $1 AND user_id = $2 AND is_system = FALSE`
	result, err := r.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
