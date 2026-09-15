package postgres

import (
	"context"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyRepo struct{ pool *pgxpool.Pool }

func NewCompanyRepo(pool *pgxpool.Pool) *CompanyRepo { return &CompanyRepo{pool: pool} }

const companyCols = `name, document, zip, street, number, complement, district, city, state, phone, email, logo, logo_width, logo_height, updated_at`

func scanCompany(row interface{ Scan(dest ...any) error }) (domain.Company, error) {
	var c domain.Company
	err := row.Scan(&c.Name, &c.Document, &c.Zip, &c.Street, &c.Number, &c.Complement, &c.District, &c.City, &c.State, &c.Phone, &c.Email, &c.Logo, &c.LogoWidth, &c.LogoHeight, &c.UpdatedAt)
	return c, err
}

func (r *CompanyRepo) Get(ctx context.Context) (domain.Company, error) {
	return scanCompany(r.pool.QueryRow(ctx, `SELECT `+companyCols+` FROM company WHERE id=1`))
}

func (r *CompanyRepo) Update(ctx context.Context, c domain.Company) (domain.Company, error) {
	return scanCompany(r.pool.QueryRow(ctx, `
		UPDATE company SET
			name=$1, document=$2, zip=$3, street=$4, number=$5, complement=$6, district=$7, city=$8, state=$9,
			phone=$10, email=$11, logo=$12, logo_width=$13, logo_height=$14, updated_at=NOW()
		WHERE id=1
		RETURNING `+companyCols,
		c.Name, c.Document, c.Zip, c.Street, c.Number, c.Complement, c.District, c.City, c.State,
		c.Phone, c.Email, c.Logo, c.LogoWidth, c.LogoHeight,
	))
}
