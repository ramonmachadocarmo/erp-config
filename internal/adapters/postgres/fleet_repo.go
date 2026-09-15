package postgres

import (
	"context"
	"errors"

	"erp/services/config-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CenterRepo struct{ pool *pgxpool.Pool }

func NewCenterRepo(pool *pgxpool.Pool) *CenterRepo { return &CenterRepo{pool: pool} }

func (r *CenterRepo) Create(ctx context.Context, c domain.DistributionCenter) (domain.DistributionCenter, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO distribution_centers (code, name, warehouse_id, zip, street, number, complement, district, city, state, lat, lng, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,TRUE)
		RETURNING id, active, created_at
	`, c.Code, c.Name, c.WarehouseID, c.Zip, c.Street, c.Number, c.Complement, c.District, c.City, c.State, c.Lat, c.Lng).
		Scan(&c.ID, &c.Active, &c.CreatedAt)
	if isUnique(err) {
		return domain.DistributionCenter{}, domain.ErrConflict
	}
	return c, err
}

func (r *CenterRepo) Update(ctx context.Context, c domain.DistributionCenter) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE distribution_centers SET code=$2, name=$3, warehouse_id=$4, zip=$5, street=$6, number=$7, complement=$8, district=$9, city=$10, state=$11, lat=$12, lng=$13, active=$14
		WHERE id=$1
	`, c.ID, c.Code, c.Name, c.WarehouseID, c.Zip, c.Street, c.Number, c.Complement, c.District, c.City, c.State, c.Lat, c.Lng, c.Active)
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

func (r *CenterRepo) Get(ctx context.Context, id string) (domain.DistributionCenter, error) {
	var c domain.DistributionCenter
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, warehouse_id::text, zip, street, number, complement, district, city, state, lat, lng, active, created_at
		FROM distribution_centers WHERE id=$1
	`, id).Scan(&c.ID, &c.Code, &c.Name, &c.WarehouseID, &c.Zip, &c.Street, &c.Number, &c.Complement, &c.District, &c.City, &c.State, &c.Lat, &c.Lng, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DistributionCenter{}, domain.ErrNotFound
	}
	return c, err
}

func (r *CenterRepo) List(ctx context.Context) ([]domain.DistributionCenter, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, warehouse_id::text, zip, street, number, complement, district, city, state, lat, lng, active, created_at
		FROM distribution_centers ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DistributionCenter
	for rows.Next() {
		var c domain.DistributionCenter
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.WarehouseID, &c.Zip, &c.Street, &c.Number, &c.Complement, &c.District, &c.City, &c.State, &c.Lat, &c.Lng, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []domain.DistributionCenter{}
	}
	return out, rows.Err()
}

type VehicleRepo struct{ pool *pgxpool.Pool }

func NewVehicleRepo(pool *pgxpool.Pool) *VehicleRepo { return &VehicleRepo{pool: pool} }

func (r *VehicleRepo) Create(ctx context.Context, v domain.DeliveryVehicle) (domain.DeliveryVehicle, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO delivery_vehicles (code, name, capacity_m3, capacity_kg, active)
		VALUES ($1,$2,$3,$4,TRUE)
		RETURNING id, active, created_at
	`, v.Code, v.Name, v.CapacityM3, v.CapacityKg).Scan(&v.ID, &v.Active, &v.CreatedAt)
	if isUnique(err) {
		return domain.DeliveryVehicle{}, domain.ErrConflict
	}
	return v, err
}

func (r *VehicleRepo) Update(ctx context.Context, v domain.DeliveryVehicle) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE delivery_vehicles SET code=$2, name=$3, capacity_m3=$4, capacity_kg=$5, active=$6 WHERE id=$1
	`, v.ID, v.Code, v.Name, v.CapacityM3, v.CapacityKg, v.Active)
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

func (r *VehicleRepo) Get(ctx context.Context, id string) (domain.DeliveryVehicle, error) {
	var v domain.DeliveryVehicle
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, capacity_m3, capacity_kg, active, created_at FROM delivery_vehicles WHERE id=$1
	`, id).Scan(&v.ID, &v.Code, &v.Name, &v.CapacityM3, &v.CapacityKg, &v.Active, &v.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DeliveryVehicle{}, domain.ErrNotFound
	}
	return v, err
}

func (r *VehicleRepo) List(ctx context.Context) ([]domain.DeliveryVehicle, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, capacity_m3, capacity_kg, active, created_at FROM delivery_vehicles ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DeliveryVehicle
	for rows.Next() {
		var v domain.DeliveryVehicle
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.CapacityM3, &v.CapacityKg, &v.Active, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if out == nil {
		out = []domain.DeliveryVehicle{}
	}
	return out, rows.Err()
}
