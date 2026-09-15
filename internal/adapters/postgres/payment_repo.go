package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MethodRepo struct{ pool *pgxpool.Pool }

func NewMethodRepo(pool *pgxpool.Pool) *MethodRepo { return &MethodRepo{pool: pool} }

func (r *MethodRepo) Create(ctx context.Context, m domain.PaymentMethod) (domain.PaymentMethod, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payment_methods (code, name, fee_percent, fee_fixed, active) VALUES ($1,$2,$3,$4,TRUE)
		RETURNING id, active, created_at
	`, m.Code, m.Name, m.FeePercent, m.FeeFixed).Scan(&m.ID, &m.Active, &m.CreatedAt)
	if isUnique(err) {
		return domain.PaymentMethod{}, domain.ErrConflict
	}
	return m, err
}

func (r *MethodRepo) Update(ctx context.Context, m domain.PaymentMethod) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE payment_methods SET code=$2, name=$3, fee_percent=$4, fee_fixed=$5, active=$6 WHERE id=$1
	`, m.ID, m.Code, m.Name, m.FeePercent, m.FeeFixed, m.Active)
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

func (r *MethodRepo) Get(ctx context.Context, id string) (domain.PaymentMethod, error) {
	var m domain.PaymentMethod
	err := r.pool.QueryRow(ctx, `SELECT id, code, name, fee_percent, fee_fixed, active, created_at FROM payment_methods WHERE id=$1`, id).
		Scan(&m.ID, &m.Code, &m.Name, &m.FeePercent, &m.FeeFixed, &m.Active, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PaymentMethod{}, domain.ErrNotFound
	}
	return m, err
}

func (r *MethodRepo) List(ctx context.Context) ([]domain.PaymentMethod, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, fee_percent, fee_fixed, active, created_at FROM payment_methods ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PaymentMethod
	for rows.Next() {
		var m domain.PaymentMethod
		if err := rows.Scan(&m.ID, &m.Code, &m.Name, &m.FeePercent, &m.FeeFixed, &m.Active, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if out == nil {
		out = []domain.PaymentMethod{}
	}
	return out, rows.Err()
}

type TermRepo struct{ pool *pgxpool.Pool }

func NewTermRepo(pool *pgxpool.Pool) *TermRepo { return &TermRepo{pool: pool} }

func (r *TermRepo) Create(ctx context.Context, t domain.PaymentTerm) (domain.PaymentTerm, error) {
	raw, err := json.Marshal(t.Installments)
	if err != nil {
		return domain.PaymentTerm{}, err
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO payment_terms (code, name, installments, active) VALUES ($1,$2,$3,TRUE)
		RETURNING id, active, created_at
	`, t.Code, t.Name, raw).Scan(&t.ID, &t.Active, &t.CreatedAt)
	if isUnique(err) {
		return domain.PaymentTerm{}, domain.ErrConflict
	}
	return t, err
}

func (r *TermRepo) Update(ctx context.Context, t domain.PaymentTerm) error {
	raw, err := json.Marshal(t.Installments)
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, `UPDATE payment_terms SET code=$2, name=$3, installments=$4, active=$5 WHERE id=$1`, t.ID, t.Code, t.Name, raw, t.Active)
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

func (r *TermRepo) Get(ctx context.Context, id string) (domain.PaymentTerm, error) {
	return scanTerm(r.pool.QueryRow(ctx, `SELECT id, code, name, installments, active, created_at FROM payment_terms WHERE id=$1`, id))
}

func (r *TermRepo) List(ctx context.Context) ([]domain.PaymentTerm, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, installments, active, created_at FROM payment_terms ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PaymentTerm
	for rows.Next() {
		t, err := scanTerm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if out == nil {
		out = []domain.PaymentTerm{}
	}
	return out, rows.Err()
}

type termScanner interface {
	Scan(dest ...any) error
}

func scanTerm(s termScanner) (domain.PaymentTerm, error) {
	var t domain.PaymentTerm
	var raw []byte
	err := s.Scan(&t.ID, &t.Code, &t.Name, &raw, &t.Active, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PaymentTerm{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PaymentTerm{}, err
	}
	if err := json.Unmarshal(raw, &t.Installments); err != nil {
		return domain.PaymentTerm{}, err
	}
	return t, nil
}
