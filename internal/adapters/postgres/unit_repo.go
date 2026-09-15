package postgres

import (
	"context"
	"errors"
	"strings"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitRepo struct {
	pool *pgxpool.Pool
}

func NewUnitRepo(pool *pgxpool.Pool) *UnitRepo {
	return &UnitRepo{pool: pool}
}

func (r *UnitRepo) Create(ctx context.Context, u domain.UnitOfMeasure) (domain.UnitOfMeasure, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO units_of_measure (code, name, symbol, active)
		VALUES ($1,$2,$3,TRUE)
		RETURNING id, active, created_at
	`, u.Code, u.Name, u.Symbol).Scan(&u.ID, &u.Active, &u.CreatedAt)
	if isUnique(err) {
		return domain.UnitOfMeasure{}, domain.ErrConflict
	}
	return u, err
}

func (r *UnitRepo) Update(ctx context.Context, u domain.UnitOfMeasure) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE units_of_measure SET code=$2, name=$3, symbol=$4, active=$5 WHERE id=$1
	`, u.ID, u.Code, u.Name, u.Symbol, u.Active)
	if err != nil {
		if isUnique(err) {
			return domain.ErrConflict
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UnitRepo) Get(ctx context.Context, id string) (domain.UnitOfMeasure, error) {
	var u domain.UnitOfMeasure
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, symbol, active, created_at FROM units_of_measure WHERE id=$1
	`, id).Scan(&u.ID, &u.Code, &u.Name, &u.Symbol, &u.Active, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.UnitOfMeasure{}, domain.ErrNotFound
	}
	return u, err
}

func (r *UnitRepo) List(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, symbol, active, created_at
		FROM units_of_measure ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.UnitOfMeasure
	for rows.Next() {
		var u domain.UnitOfMeasure
		if err := rows.Scan(&u.ID, &u.Code, &u.Name, &u.Symbol, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if out == nil {
		out = []domain.UnitOfMeasure{}
	}
	return out, rows.Err()
}

func (r *UnitRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM units_of_measure WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "23505" || strings.Contains(err.Error(), "duplicate"))
}
