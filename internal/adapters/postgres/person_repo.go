package postgres

import (
	"context"
	"errors"
	"time"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const personCols = `p.id, p.kind, p.document, p.name, p.phone, p.birth_date, p.gender, p.company_name, p.responsible_name, p.created_at`

type PersonRepo struct {
	pool *pgxpool.Pool
}

func NewPersonRepo(pool *pgxpool.Pool) *PersonRepo {
	return &PersonRepo{pool: pool}
}

func (r *PersonRepo) Create(ctx context.Context, p domain.Person) (domain.Person, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Person{}, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `
		INSERT INTO people (kind, document, name, phone, birth_date, gender, company_name, responsible_name)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at
	`, p.Kind, p.Document, p.Name, p.Phone, nullDate(p.BirthDate), p.Gender, p.CompanyName, p.ResponsibleName).Scan(&p.ID, &p.CreatedAt)
	if isUnique(err) {
		return domain.Person{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Person{}, err
	}
	if err := replaceAddresses(ctx, tx, p.ID, p.Addresses); err != nil {
		return domain.Person{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Person{}, err
	}
	return r.Get(ctx, p.ID)
}

func (r *PersonRepo) Update(ctx context.Context, p domain.Person) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
		UPDATE people SET kind=$2, document=$3, name=$4, phone=$5, birth_date=$6, gender=$7, company_name=$8, responsible_name=$9
		WHERE id=$1
	`, p.ID, p.Kind, p.Document, p.Name, p.Phone, nullDate(p.BirthDate), p.Gender, p.CompanyName, p.ResponsibleName)
	if err != nil {
		if isUnique(err) {
			return domain.ErrConflict
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err := replaceAddresses(ctx, tx, p.ID, p.Addresses); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PersonRepo) Get(ctx context.Context, id string) (domain.Person, error) {
	p, err := scanPerson(r.pool.QueryRow(ctx, `SELECT `+personCols+` FROM people p WHERE p.id=$1`, id))
	if err != nil {
		return domain.Person{}, err
	}
	return withAddresses(ctx, r.pool, p)
}

func (r *PersonRepo) GetByDocument(ctx context.Context, document string) (domain.Person, error) {
	p, err := scanPerson(r.pool.QueryRow(ctx, `SELECT `+personCols+` FROM people p WHERE p.document=$1`, document))
	if err != nil {
		return domain.Person{}, err
	}
	return withAddresses(ctx, r.pool, p)
}

type RoleRepo struct {
	pool  *pgxpool.Pool
	table string
}

func NewCustomerRepo(pool *pgxpool.Pool) *RoleRepo {
	return &RoleRepo{pool: pool, table: "customers"}
}

func NewSupplierRepo(pool *pgxpool.Pool) *RoleRepo {
	return &RoleRepo{pool: pool, table: "suppliers"}
}

func (r *RoleRepo) Link(ctx context.Context, personID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO `+r.table+` (id) VALUES ($1)`, personID)
	if isUnique(err) {
		return domain.ErrConflict
	}
	return err
}

func (r *RoleRepo) Unlink(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM `+r.table+` WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RoleRepo) Get(ctx context.Context, id string) (domain.Person, error) {
	p, err := scanPerson(r.pool.QueryRow(ctx, `
		SELECT `+personCols+` FROM people p INNER JOIN `+r.table+` r ON r.id = p.id WHERE p.id=$1
	`, id))
	if err != nil {
		return domain.Person{}, err
	}
	return withAddresses(ctx, r.pool, p)
}

func (r *RoleRepo) List(ctx context.Context) ([]domain.Person, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+personCols+` FROM people p INNER JOIN `+r.table+` r ON r.id = p.id ORDER BY p.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Person
	var ids []string
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
		ids = append(ids, p.ID)
	}
	if out == nil {
		return []domain.Person{}, rows.Err()
	}
	byID, err := loadAddressMap(ctx, r.pool, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Addresses = byID[out[i].ID]
		if out[i].Addresses == nil {
			out[i].Addresses = []domain.Address{}
		}
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPerson(row rowScanner) (domain.Person, error) {
	var p domain.Person
	var birth *time.Time
	err := row.Scan(&p.ID, &p.Kind, &p.Document, &p.Name, &p.Phone, &birth, &p.Gender, &p.CompanyName, &p.ResponsibleName, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Person{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Person{}, err
	}
	if birth != nil {
		p.BirthDate = birth.Format("2006-01-02")
	}
	p.Addresses = []domain.Address{}
	return p, nil
}

func withAddresses(ctx context.Context, pool *pgxpool.Pool, p domain.Person) (domain.Person, error) {
	m, err := loadAddressMap(ctx, pool, []string{p.ID})
	if err != nil {
		return domain.Person{}, err
	}
	p.Addresses = m[p.ID]
	if p.Addresses == nil {
		p.Addresses = []domain.Address{}
	}
	return p, nil
}

func loadAddressMap(ctx context.Context, pool *pgxpool.Pool, ids []string) (map[string][]domain.Address, error) {
	out := map[string][]domain.Address{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := pool.Query(ctx, `
		SELECT id, person_id, alias, zip, street, number, complement, district, city, state, lat, lng
		FROM person_addresses WHERE person_id = ANY($1::uuid[]) ORDER BY created_at
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var a domain.Address
		var personID string
		var lat, lng *float64
		if err := rows.Scan(&a.ID, &personID, &a.Alias, &a.Zip, &a.Street, &a.Number, &a.Complement, &a.District, &a.City, &a.State, &lat, &lng); err != nil {
			return nil, err
		}
		a.Lat, a.Lng = lat, lng
		out[personID] = append(out[personID], a)
	}
	return out, rows.Err()
}

func replaceAddresses(ctx context.Context, tx pgx.Tx, personID string, addrs []domain.Address) error {
	if _, err := tx.Exec(ctx, `DELETE FROM person_addresses WHERE person_id=$1`, personID); err != nil {
		return err
	}
	for _, a := range addrs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO person_addresses (person_id, alias, zip, street, number, complement, district, city, state, lat, lng)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		`, personID, a.Alias, a.Zip, a.Street, a.Number, a.Complement, a.District, a.City, a.State, a.Lat, a.Lng); err != nil {
			return err
		}
	}
	return nil
}

func nullDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}
