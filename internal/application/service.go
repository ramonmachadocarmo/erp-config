package application

import (
	"context"
	"errors"
	"strings"

	"erp/pkg/codes"
	"erp/services/config-service/internal/domain"
)

type Service struct {
	units      domain.UnitRepository
	people     domain.PersonRepository
	customers  domain.RoleRepository
	suppliers  domain.RoleRepository
	settings   domain.SettingRepository
	methods    domain.PaymentMethodRepository
	terms      domain.PaymentTermRepository
	centers    domain.CenterRepository
	vehicles   domain.VehicleRepository
	company    domain.CompanyRepository
	geo        domain.PostalLookup
	unitSeq    codes.Sequence
	methodSeq  codes.Sequence
	termSeq    codes.Sequence
	centerSeq  codes.Sequence
	vehicleSeq codes.Sequence
}

func New(units domain.UnitRepository, people domain.PersonRepository, customers domain.RoleRepository, suppliers domain.RoleRepository, settings domain.SettingRepository, methods domain.PaymentMethodRepository, terms domain.PaymentTermRepository, centers domain.CenterRepository, vehicles domain.VehicleRepository, company domain.CompanyRepository, geo domain.PostalLookup, unitSeq, methodSeq, termSeq, centerSeq, vehicleSeq codes.Sequence) *Service {
	return &Service{units: units, people: people, customers: customers, suppliers: suppliers, settings: settings, methods: methods, terms: terms, centers: centers, vehicles: vehicles, company: company, geo: geo, unitSeq: unitSeq, methodSeq: methodSeq, termSeq: termSeq, centerSeq: centerSeq, vehicleSeq: vehicleSeq}
}

func (s *Service) CreateUnit(ctx context.Context, u domain.UnitOfMeasure) (domain.UnitOfMeasure, error) {
	u.Code = strings.ToUpper(strings.TrimSpace(u.Code))
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return domain.UnitOfMeasure{}, domain.ErrInvalid
	}
	code, err := codes.Assign(ctx, u.Code, s.unitSeq)
	if err != nil {
		return domain.UnitOfMeasure{}, err
	}
	u.Code = code
	if u.Symbol == "" {
		u.Symbol = strings.ToLower(u.Code)
	}
	u.Active = true
	return s.units.Create(ctx, u)
}

func (s *Service) UpdateUnit(ctx context.Context, u domain.UnitOfMeasure) error {
	u.Code = strings.ToUpper(strings.TrimSpace(u.Code))
	u.Name = strings.TrimSpace(u.Name)
	return s.units.Update(ctx, u)
}

func (s *Service) GetUnit(ctx context.Context, id string) (domain.UnitOfMeasure, error) {
	return s.units.Get(ctx, id)
}

func (s *Service) ListUnits(ctx context.Context) ([]domain.UnitOfMeasure, error) {
	return s.units.List(ctx)
}

func (s *Service) DeleteUnit(ctx context.Context, id string) error {
	return domain.ErrInUse
}

func (s *Service) CreateCustomer(ctx context.Context, p domain.Person) (domain.Person, error) {
	return s.ensureRole(ctx, p, s.customers)
}

func (s *Service) UpdateCustomer(ctx context.Context, p domain.Person) error {
	return s.updateRole(ctx, p, s.customers)
}

func (s *Service) GetCustomer(ctx context.Context, id string) (domain.Person, error) {
	return s.customers.Get(ctx, id)
}

func (s *Service) ListCustomers(ctx context.Context) ([]domain.Person, error) {
	return s.customers.List(ctx)
}

func (s *Service) DeleteCustomer(ctx context.Context, id string) error {
	return s.customers.Unlink(ctx, id)
}

func (s *Service) CreateSupplier(ctx context.Context, p domain.Person) (domain.Person, error) {
	return s.ensureRole(ctx, p, s.suppliers)
}

func (s *Service) UpdateSupplier(ctx context.Context, p domain.Person) error {
	return s.updateRole(ctx, p, s.suppliers)
}

func (s *Service) GetSupplier(ctx context.Context, id string) (domain.Person, error) {
	return s.suppliers.Get(ctx, id)
}

func (s *Service) ListSuppliers(ctx context.Context) ([]domain.Person, error) {
	return s.suppliers.List(ctx)
}

func (s *Service) DeleteSupplier(ctx context.Context, id string) error {
	return s.suppliers.Unlink(ctx, id)
}

func (s *Service) ListSettings(ctx context.Context) ([]domain.Setting, error) {
	return s.settings.List(ctx)
}

func allowedSetting(key string) bool {
	return key == domain.SettingPurchaseQuoteRequired || key == domain.SettingDefaultWarehouse
}

func (s *Service) GetSetting(ctx context.Context, key string) (domain.Setting, error) {
	if !allowedSetting(key) {
		return domain.Setting{}, domain.ErrNotFound
	}
	st, err := s.settings.Get(ctx, key)
	if errors.Is(err, domain.ErrNotFound) {
		def := ""
		if key == domain.SettingPurchaseQuoteRequired {
			def = "false"
		}
		return domain.Setting{Key: key, Value: def}, nil
	}
	return st, err
}

func (s *Service) PutSetting(ctx context.Context, st domain.Setting) error {
	if !allowedSetting(st.Key) {
		return domain.ErrInvalid
	}
	if st.Key == domain.SettingPurchaseQuoteRequired {
		v := strings.ToLower(strings.TrimSpace(st.Value))
		if v != "true" && v != "false" {
			return domain.ErrInvalid
		}
		st.Value = v
		return s.settings.Upsert(ctx, st)
	}
	st.Value = strings.TrimSpace(st.Value)
	return s.settings.Upsert(ctx, st)
}

func (s *Service) ensureRole(ctx context.Context, p domain.Person, roles domain.RoleRepository) (domain.Person, error) {
	p.Normalize()
	if err := p.Validate(); err != nil {
		return domain.Person{}, err
	}
	// A blank document can't be used to find "the same person" — unlike a real CPF/CNPJ, it's
	// not unique (see migration 010), so looking it up would non-deterministically match some
	// other, unrelated customer with no document on file and silently overwrite their record.
	// Always create a new person instead when there's no document to match on.
	var existing domain.Person
	err := domain.ErrNotFound
	if p.Document != "" {
		existing, err = s.people.GetByDocument(ctx, p.Document)
	}
	if err == nil {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
		if err := s.people.Update(ctx, p); err != nil {
			return domain.Person{}, err
		}
	} else if errors.Is(err, domain.ErrNotFound) {
		created, err := s.people.Create(ctx, p)
		if err != nil {
			return domain.Person{}, err
		}
		p = created
	} else {
		return domain.Person{}, err
	}
	if err := roles.Link(ctx, p.ID); err != nil {
		return domain.Person{}, err
	}
	return p, nil
}

func (s *Service) updateRole(ctx context.Context, p domain.Person, roles domain.RoleRepository) error {
	if _, err := roles.Get(ctx, p.ID); err != nil {
		return err
	}
	p.Normalize()
	if err := p.Validate(); err != nil {
		return err
	}
	return s.people.Update(ctx, p)
}

func (s *Service) CreatePaymentMethod(ctx context.Context, m domain.PaymentMethod) (domain.PaymentMethod, error) {
	m.Code = strings.ToUpper(strings.TrimSpace(m.Code))
	m.Name = strings.TrimSpace(m.Name)
	if m.Name == "" || m.FeePercent < 0 || m.FeeFixed < 0 {
		return domain.PaymentMethod{}, domain.ErrInvalid
	}
	code, err := codes.Assign(ctx, m.Code, s.methodSeq)
	if err != nil {
		return domain.PaymentMethod{}, err
	}
	m.Code = code
	m.Active = true
	return s.methods.Create(ctx, m)
}

func (s *Service) UpdatePaymentMethod(ctx context.Context, m domain.PaymentMethod) error {
	m.Code = strings.ToUpper(strings.TrimSpace(m.Code))
	m.Name = strings.TrimSpace(m.Name)
	if m.Code == "" || m.Name == "" || m.FeePercent < 0 || m.FeeFixed < 0 {
		return domain.ErrInvalid
	}
	return s.methods.Update(ctx, m)
}

func (s *Service) GetPaymentMethod(ctx context.Context, id string) (domain.PaymentMethod, error) {
	return s.methods.Get(ctx, id)
}

func (s *Service) ListPaymentMethods(ctx context.Context) ([]domain.PaymentMethod, error) {
	return s.methods.List(ctx)
}

func (s *Service) CreatePaymentTerm(ctx context.Context, t domain.PaymentTerm) (domain.PaymentTerm, error) {
	t.Code = strings.ToUpper(strings.TrimSpace(t.Code))
	t.Name = strings.TrimSpace(t.Name)
	code, err := codes.Assign(ctx, t.Code, s.termSeq)
	if err != nil {
		return domain.PaymentTerm{}, err
	}
	t.Code = code
	if err := t.Validate(); err != nil {
		return domain.PaymentTerm{}, err
	}
	t.Active = true
	return s.terms.Create(ctx, t)
}

func (s *Service) UpdatePaymentTerm(ctx context.Context, t domain.PaymentTerm) error {
	t.Code = strings.ToUpper(strings.TrimSpace(t.Code))
	t.Name = strings.TrimSpace(t.Name)
	if err := t.Validate(); err != nil {
		return err
	}
	return s.terms.Update(ctx, t)
}

func (s *Service) GetPaymentTerm(ctx context.Context, id string) (domain.PaymentTerm, error) {
	return s.terms.Get(ctx, id)
}

func (s *Service) ListPaymentTerms(ctx context.Context) ([]domain.PaymentTerm, error) {
	return s.terms.List(ctx)
}

func (s *Service) LookupCEP(ctx context.Context, cep string) (domain.Address, error) {
	return s.geo.ByCEP(ctx, cep)
}

func (s *Service) SearchCEP(ctx context.Context, state, city, street, district string) ([]domain.Address, error) {
	return s.geo.SearchCEP(ctx, state, city, street, district)
}

func (s *Service) LookupGeo(ctx context.Context, lat, lng float64) (domain.Address, error) {
	return s.geo.ByGeo(ctx, lat, lng)
}

func (s *Service) SearchGeo(ctx context.Context, query string) (domain.Address, error) {
	return s.geo.Search(ctx, query)
}

func (s *Service) SearchPlace(ctx context.Context, street, number, district, city, state, zip string) (domain.Address, error) {
	return s.geo.SearchPlace(ctx, street, number, district, city, state, zip)
}

func (s *Service) CreateCenter(ctx context.Context, c domain.DistributionCenter) (domain.DistributionCenter, error) {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	c.Name = strings.TrimSpace(c.Name)
	c.WarehouseID = strings.TrimSpace(c.WarehouseID)
	c.Zip = strings.TrimSpace(c.Zip)
	c.Street = strings.TrimSpace(c.Street)
	c.Number = strings.TrimSpace(c.Number)
	c.Complement = strings.TrimSpace(c.Complement)
	c.District = strings.TrimSpace(c.District)
	c.City = strings.TrimSpace(c.City)
	c.State = strings.ToUpper(strings.TrimSpace(c.State))
	if err := c.Validate(); err != nil {
		return domain.DistributionCenter{}, err
	}
	code, err := codes.Assign(ctx, c.Code, s.centerSeq)
	if err != nil {
		return domain.DistributionCenter{}, err
	}
	c.Code = code
	c.Active = true
	return s.centers.Create(ctx, c)
}

func (s *Service) UpdateCenter(ctx context.Context, c domain.DistributionCenter) error {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	c.Name = strings.TrimSpace(c.Name)
	c.WarehouseID = strings.TrimSpace(c.WarehouseID)
	c.State = strings.ToUpper(strings.TrimSpace(c.State))
	if err := c.Validate(); err != nil {
		return err
	}
	return s.centers.Update(ctx, c)
}

func (s *Service) GetCenter(ctx context.Context, id string) (domain.DistributionCenter, error) {
	return s.centers.Get(ctx, id)
}

func (s *Service) ListCenters(ctx context.Context) ([]domain.DistributionCenter, error) {
	return s.centers.List(ctx)
}

func (s *Service) CreateVehicle(ctx context.Context, v domain.DeliveryVehicle) (domain.DeliveryVehicle, error) {
	v.Code = strings.ToUpper(strings.TrimSpace(v.Code))
	v.Name = strings.TrimSpace(v.Name)
	if err := v.Validate(); err != nil {
		return domain.DeliveryVehicle{}, err
	}
	code, err := codes.Assign(ctx, v.Code, s.vehicleSeq)
	if err != nil {
		return domain.DeliveryVehicle{}, err
	}
	v.Code = code
	v.Active = true
	return s.vehicles.Create(ctx, v)
}

func (s *Service) UpdateVehicle(ctx context.Context, v domain.DeliveryVehicle) error {
	v.Code = strings.ToUpper(strings.TrimSpace(v.Code))
	v.Name = strings.TrimSpace(v.Name)
	if err := v.Validate(); err != nil {
		return err
	}
	return s.vehicles.Update(ctx, v)
}

func (s *Service) GetVehicle(ctx context.Context, id string) (domain.DeliveryVehicle, error) {
	return s.vehicles.Get(ctx, id)
}

func (s *Service) ListVehicles(ctx context.Context) ([]domain.DeliveryVehicle, error) {
	return s.vehicles.List(ctx)
}

func (s *Service) GetCompany(ctx context.Context) (domain.Company, error) {
	return s.company.Get(ctx)
}

func (s *Service) UpdateCompany(ctx context.Context, c domain.Company) (domain.Company, error) {
	c.Name = strings.TrimSpace(c.Name)
	c.Document = strings.TrimSpace(c.Document)
	c.Zip = strings.TrimSpace(c.Zip)
	c.Street = strings.TrimSpace(c.Street)
	c.Number = strings.TrimSpace(c.Number)
	c.Complement = strings.TrimSpace(c.Complement)
	c.District = strings.TrimSpace(c.District)
	c.City = strings.TrimSpace(c.City)
	c.State = strings.ToUpper(strings.TrimSpace(c.State))
	c.Phone = strings.TrimSpace(c.Phone)
	c.Email = strings.TrimSpace(c.Email)
	return s.company.Update(ctx, c)
}
