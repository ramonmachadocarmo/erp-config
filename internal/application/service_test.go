package application

import (
	"context"
	"fmt"
	"testing"

	"erp/services/config-service/internal/domain"
)

type memPeople struct {
	byID   map[string]domain.Person
	nextID int
}

func (m *memPeople) Create(_ context.Context, p domain.Person) (domain.Person, error) {
	if m.byID == nil {
		m.byID = map[string]domain.Person{}
	}
	m.nextID++
	p.ID = fmt.Sprintf("p%d", m.nextID)
	m.byID[p.ID] = p
	return p, nil
}

func (m *memPeople) Update(_ context.Context, p domain.Person) error {
	if _, ok := m.byID[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.byID[p.ID] = p
	return nil
}

func (m *memPeople) Get(_ context.Context, id string) (domain.Person, error) {
	p, ok := m.byID[id]
	if !ok {
		return domain.Person{}, domain.ErrNotFound
	}
	return p, nil
}

func (m *memPeople) GetByDocument(_ context.Context, document string) (domain.Person, error) {
	for _, p := range m.byID {
		if p.Document == document {
			return p, nil
		}
	}
	return domain.Person{}, domain.ErrNotFound
}

type memSettings struct{ m map[string]domain.Setting }

func (s *memSettings) Get(_ context.Context, key string) (domain.Setting, error) {
	st, ok := s.m[key]
	if !ok {
		return domain.Setting{}, domain.ErrNotFound
	}
	return st, nil
}
func (s *memSettings) List(context.Context) ([]domain.Setting, error) { return nil, nil }
func (s *memSettings) Upsert(_ context.Context, st domain.Setting) error {
	if s.m == nil {
		s.m = map[string]domain.Setting{}
	}
	s.m[st.Key] = st
	return nil
}

func cfgSvc() (*Service, *memSettings) {
	st := &memSettings{m: map[string]domain.Setting{}}
	return New(nil, nil, nil, nil, st, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil), st
}

type memRoles struct{ linked map[string]bool }

func (r *memRoles) Link(_ context.Context, id string) error {
	if r.linked == nil {
		r.linked = map[string]bool{}
	}
	if r.linked[id] {
		return domain.ErrConflict
	}
	r.linked[id] = true
	return nil
}

func (r *memRoles) Unlink(_ context.Context, id string) error {
	if !r.linked[id] {
		return domain.ErrNotFound
	}
	delete(r.linked, id)
	return nil
}

func (r *memRoles) Get(_ context.Context, id string) (domain.Person, error) {
	if !r.linked[id] {
		return domain.Person{}, domain.ErrNotFound
	}
	return domain.Person{ID: id}, nil
}

func (r *memRoles) List(context.Context) ([]domain.Person, error) { return nil, nil }

type memMethods struct {
	byID   map[string]domain.PaymentMethod
	nextID int
}

func (m *memMethods) Create(_ context.Context, pm domain.PaymentMethod) (domain.PaymentMethod, error) {
	if m.byID == nil {
		m.byID = map[string]domain.PaymentMethod{}
	}
	m.nextID++
	pm.ID = fmt.Sprintf("m%d", m.nextID)
	m.byID[pm.ID] = pm
	return pm, nil
}

func (m *memMethods) Update(_ context.Context, pm domain.PaymentMethod) error {
	if _, ok := m.byID[pm.ID]; !ok {
		return domain.ErrNotFound
	}
	m.byID[pm.ID] = pm
	return nil
}

func (m *memMethods) Get(_ context.Context, id string) (domain.PaymentMethod, error) {
	pm, ok := m.byID[id]
	if !ok {
		return domain.PaymentMethod{}, domain.ErrNotFound
	}
	return pm, nil
}

func (m *memMethods) List(context.Context) ([]domain.PaymentMethod, error) { return nil, nil }

type seqStub struct{ n int64 }

func (s *seqStub) Next(context.Context) (int64, error)    { s.n++; return s.n, nil }
func (s *seqStub) EnsureMin(context.Context, int64) error { return nil }

func TestGetSettingDefaults(t *testing.T) {
	svc, _ := cfgSvc()
	q, err := svc.GetSetting(context.Background(), domain.SettingPurchaseQuoteRequired)
	if err != nil || q.Value != "false" {
		t.Fatalf("%v %+v", err, q)
	}
	w, err := svc.GetSetting(context.Background(), domain.SettingDefaultWarehouse)
	if err != nil || w.Value != "" {
		t.Fatalf("%v %+v", err, w)
	}
}

func TestPutSettingWhitelist(t *testing.T) {
	svc, _ := cfgSvc()
	if err := svc.PutSetting(context.Background(), domain.Setting{Key: "unknown", Value: "x"}); err != domain.ErrInvalid {
		t.Fatalf("%v", err)
	}
}

func TestPutQuoteRequired(t *testing.T) {
	svc, st := cfgSvc()
	if err := svc.PutSetting(context.Background(), domain.Setting{Key: domain.SettingPurchaseQuoteRequired, Value: "TRUE"}); err != nil {
		t.Fatal(err)
	}
	if st.m[domain.SettingPurchaseQuoteRequired].Value != "true" {
		t.Fatalf("%+v", st.m)
	}
	if err := svc.PutSetting(context.Background(), domain.Setting{Key: domain.SettingPurchaseQuoteRequired, Value: "maybe"}); err != domain.ErrInvalid {
		t.Fatalf("%v", err)
	}
}

func TestPutDefaultWarehouse(t *testing.T) {
	svc, st := cfgSvc()
	if err := svc.PutSetting(context.Background(), domain.Setting{Key: domain.SettingDefaultWarehouse, Value: "  w1  "}); err != nil {
		t.Fatal(err)
	}
	if st.m[domain.SettingDefaultWarehouse].Value != "w1" {
		t.Fatalf("%+v", st.m)
	}
	if err := svc.PutSetting(context.Background(), domain.Setting{Key: domain.SettingDefaultWarehouse, Value: ""}); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteCustomerUnlinksTheRoleNotThePerson(t *testing.T) {
	customers := &memRoles{linked: map[string]bool{"p1": true}}
	svc := New(nil, nil, customers, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if err := svc.DeleteCustomer(context.Background(), "p1"); err != nil {
		t.Fatalf("DeleteCustomer: %v", err)
	}
	if customers.linked["p1"] {
		t.Fatalf("expected p1 unlinked from customers")
	}
	if _, err := svc.GetCustomer(context.Background(), "p1"); err != domain.ErrNotFound {
		t.Fatalf("GetCustomer after delete = %v, want ErrNotFound", err)
	}
}

func TestDeleteCustomerNotFound(t *testing.T) {
	customers := &memRoles{}
	svc := New(nil, nil, customers, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	if err := svc.DeleteCustomer(context.Background(), "missing"); err != domain.ErrNotFound {
		t.Fatalf("DeleteCustomer = %v, want ErrNotFound", err)
	}
}

func TestCreateCustomerWithBlankDocumentNeverMerges(t *testing.T) {
	people := &memPeople{}
	customers := &memRoles{}
	svc := New(nil, people, customers, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	a, err := svc.CreateCustomer(context.Background(), domain.Person{Kind: domain.KindPF, Name: "A", Phone: "1"})
	if err != nil {
		t.Fatalf("create A: %v", err)
	}
	b, err := svc.CreateCustomer(context.Background(), domain.Person{Kind: domain.KindPF, Name: "B", Phone: "2"})
	if err != nil {
		t.Fatalf("create B: %v", err)
	}
	if a.ID == b.ID {
		t.Fatalf("two blank-document customers were merged into one person: %+v / %+v", a, b)
	}
	if a.Name != "A" || b.Name != "B" {
		t.Fatalf("a blank-document customer's data got overwritten by another: %+v / %+v", a, b)
	}
}

func TestCreateCustomerWithSameDocumentStillMerges(t *testing.T) {
	people := &memPeople{}
	customers := &memRoles{}
	suppliers := &memRoles{}
	svc := New(nil, people, customers, suppliers, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	a, err := svc.CreateCustomer(context.Background(), domain.Person{Kind: domain.KindPF, Document: "111", Name: "A", Phone: "1"})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	b, err := svc.CreateSupplier(context.Background(), domain.Person{Kind: domain.KindPF, Document: "111", Name: "A updated", Phone: "2"})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}
	if a.ID != b.ID {
		t.Fatalf("same real document should resolve to the same person: %+v / %+v", a, b)
	}
	if !customers.linked[a.ID] || !suppliers.linked[a.ID] {
		t.Fatalf("expected both roles linked to %s: customers=%v suppliers=%v", a.ID, customers.linked, suppliers.linked)
	}
}

func TestCreatePaymentMethodStoresFees(t *testing.T) {
	methods := &memMethods{}
	svc := New(nil, nil, nil, nil, nil, methods, nil, nil, nil, nil, nil, nil, &seqStub{}, nil, nil, nil)

	m, err := svc.CreatePaymentMethod(context.Background(), domain.PaymentMethod{Name: "Cartão", FeePercent: 3.5, FeeFixed: 0.5})
	if err != nil {
		t.Fatalf("CreatePaymentMethod: %v", err)
	}
	if m.FeePercent != 3.5 || m.FeeFixed != 0.5 {
		t.Fatalf("fees not preserved: %+v", m)
	}
	got, err := svc.GetPaymentMethod(context.Background(), m.ID)
	if err != nil || got.FeePercent != 3.5 || got.FeeFixed != 0.5 {
		t.Fatalf("fees not round-tripped: %v %+v", err, got)
	}
}

func TestCreatePaymentMethodRejectsNegativeFees(t *testing.T) {
	methods := &memMethods{}
	svc := New(nil, nil, nil, nil, nil, methods, nil, nil, nil, nil, nil, nil, &seqStub{}, nil, nil, nil)

	if _, err := svc.CreatePaymentMethod(context.Background(), domain.PaymentMethod{Name: "X", FeePercent: -1}); err != domain.ErrInvalid {
		t.Fatalf("fee_percent: %v", err)
	}
	if _, err := svc.CreatePaymentMethod(context.Background(), domain.PaymentMethod{Name: "X", FeeFixed: -1}); err != domain.ErrInvalid {
		t.Fatalf("fee_fixed: %v", err)
	}
}

func TestUpdatePaymentMethodRejectsNegativeFees(t *testing.T) {
	methods := &memMethods{}
	svc := New(nil, nil, nil, nil, nil, methods, nil, nil, nil, nil, nil, nil, &seqStub{}, nil, nil, nil)
	m, err := svc.CreatePaymentMethod(context.Background(), domain.PaymentMethod{Name: "Cartão"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	m.FeePercent = -2
	if err := svc.UpdatePaymentMethod(context.Background(), m); err != domain.ErrInvalid {
		t.Fatalf("%v", err)
	}
}
