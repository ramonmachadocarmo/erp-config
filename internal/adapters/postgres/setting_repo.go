package postgres

import (
	"context"
	"errors"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SettingRepo struct {
	pool *pgxpool.Pool
}

func NewSettingRepo(pool *pgxpool.Pool) *SettingRepo {
	return &SettingRepo{pool: pool}
}

func (r *SettingRepo) Get(ctx context.Context, key string) (domain.Setting, error) {
	var s domain.Setting
	err := r.pool.QueryRow(ctx, `SELECT key, value FROM settings WHERE key=$1`, key).Scan(&s.Key, &s.Value)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Setting{}, domain.ErrNotFound
	}
	return s, err
}

func (r *SettingRepo) List(ctx context.Context) ([]domain.Setting, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Setting
	for rows.Next() {
		var s domain.Setting
		if err := rows.Scan(&s.Key, &s.Value); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if out == nil {
		out = []domain.Setting{}
	}
	return out, rows.Err()
}

func (r *SettingRepo) Upsert(ctx context.Context, s domain.Setting) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settings (key, value) VALUES ($1,$2)
		ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=NOW()
	`, s.Key, s.Value)
	return err
}
