package domain

import (
	"context"
	"errors"
	"math"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
	ErrInvalid  = errors.New("invalid data")
	ErrInUse    = errors.New("cannot delete: record is in use")
)

const (
	SettingPurchaseQuoteRequired = "purchase_quote_required"
	SettingDefaultWarehouse      = "default_warehouse_id"
)

type UnitOfMeasure struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type UnitRepository interface {
	Create(ctx context.Context, u UnitOfMeasure) (UnitOfMeasure, error)
	Update(ctx context.Context, u UnitOfMeasure) error
	Get(ctx context.Context, id string) (UnitOfMeasure, error)
	List(ctx context.Context) ([]UnitOfMeasure, error)
	Delete(ctx context.Context, id string) error
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SettingRepository interface {
	Get(ctx context.Context, key string) (Setting, error)
	List(ctx context.Context) ([]Setting, error)
	Upsert(ctx context.Context, s Setting) error
}

type PaymentMethod struct {
	ID     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	// FeePercent and FeeFixed are what the acquirer/processor charges the business for this
	// method (e.g. a card gateway's percentage cut plus a fixed per-transaction fee) — not
	// charged to the customer, just tracked so margin calculations can account for it.
	FeePercent float64   `json:"fee_percent"`
	FeeFixed   float64   `json:"fee_fixed"`
	CreatedAt  time.Time `json:"created_at"`
}

type InstallmentSpec struct {
	Days    int     `json:"days"`
	Percent float64 `json:"percent"`
}

type PaymentTerm struct {
	ID           string            `json:"id"`
	Code         string            `json:"code"`
	Name         string            `json:"name"`
	Installments []InstallmentSpec `json:"installments"`
	Active       bool              `json:"active"`
	CreatedAt    time.Time         `json:"created_at"`
}

type PaymentMethodRepository interface {
	Create(ctx context.Context, m PaymentMethod) (PaymentMethod, error)
	Update(ctx context.Context, m PaymentMethod) error
	Get(ctx context.Context, id string) (PaymentMethod, error)
	List(ctx context.Context) ([]PaymentMethod, error)
}

type PaymentTermRepository interface {
	Create(ctx context.Context, t PaymentTerm) (PaymentTerm, error)
	Update(ctx context.Context, t PaymentTerm) error
	Get(ctx context.Context, id string) (PaymentTerm, error)
	List(ctx context.Context) ([]PaymentTerm, error)
}

type DistributionCenter struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	WarehouseID string    `json:"warehouse_id"`
	Zip         string    `json:"zip"`
	Street      string    `json:"street"`
	Number      string    `json:"number"`
	Complement  string    `json:"complement"`
	District    string    `json:"district"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c DistributionCenter) Validate() error {
	if c.Name == "" || c.WarehouseID == "" || c.Lat == 0 && c.Lng == 0 {
		return ErrInvalid
	}
	return nil
}

type DeliveryVehicle struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	CapacityM3 float64   `json:"capacity_m3"`
	CapacityKg float64   `json:"capacity_kg"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

func (v DeliveryVehicle) Validate() error {
	if v.Name == "" || v.CapacityM3 <= 0 || v.CapacityKg <= 0 {
		return ErrInvalid
	}
	return nil
}

type CenterRepository interface {
	Create(ctx context.Context, c DistributionCenter) (DistributionCenter, error)
	Update(ctx context.Context, c DistributionCenter) error
	Get(ctx context.Context, id string) (DistributionCenter, error)
	List(ctx context.Context) ([]DistributionCenter, error)
}

type Company struct {
	Name       string    `json:"name"`
	Document   string    `json:"document"`
	Zip        string    `json:"zip"`
	Street     string    `json:"street"`
	Number     string    `json:"number"`
	Complement string    `json:"complement"`
	District   string    `json:"district"`
	City       string    `json:"city"`
	State      string    `json:"state"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Logo       string    `json:"logo"`
	LogoWidth  int       `json:"logo_width"`
	LogoHeight int       `json:"logo_height"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CompanyRepository interface {
	Get(ctx context.Context) (Company, error)
	Update(ctx context.Context, c Company) (Company, error)
}

type VehicleRepository interface {
	Create(ctx context.Context, v DeliveryVehicle) (DeliveryVehicle, error)
	Update(ctx context.Context, v DeliveryVehicle) error
	Get(ctx context.Context, id string) (DeliveryVehicle, error)
	List(ctx context.Context) ([]DeliveryVehicle, error)
}

func (t PaymentTerm) Validate() error {
	if t.Code == "" || t.Name == "" || len(t.Installments) == 0 {
		return ErrInvalid
	}
	var sum float64
	for _, i := range t.Installments {
		if i.Days < 0 || i.Percent <= 0 {
			return ErrInvalid
		}
		sum += i.Percent
	}
	if math.Abs(sum-100) > 0.05 {
		return ErrInvalid
	}
	return nil
}
