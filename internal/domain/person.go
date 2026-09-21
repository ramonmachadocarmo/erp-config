package domain

import (
	"context"
	"strings"
	"time"
)

const (
	KindPF = "PF"
	KindPJ = "PJ"
)

type Address struct {
	ID         string   `json:"id,omitempty"`
	Alias      string   `json:"alias"`
	Zip        string   `json:"zip"`
	Street     string   `json:"street"`
	Number     string   `json:"number"`
	Complement string   `json:"complement"`
	District   string   `json:"district"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	Lat        *float64 `json:"lat,omitempty"`
	Lng        *float64 `json:"lng,omitempty"`
}

func (a *Address) Normalize() {
	a.Alias = strings.TrimSpace(a.Alias)
	a.Zip = strings.TrimSpace(a.Zip)
	a.Street = strings.TrimSpace(a.Street)
	a.Number = strings.TrimSpace(a.Number)
	a.Complement = strings.TrimSpace(a.Complement)
	a.District = strings.TrimSpace(a.District)
	a.City = strings.TrimSpace(a.City)
	a.State = strings.ToUpper(strings.TrimSpace(a.State))
}

func (a Address) Empty() bool {
	return a.Alias == "" && a.Zip == "" && a.Street == "" && a.Number == "" && a.Complement == "" && a.District == "" && a.City == "" && a.State == ""
}

type Person struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"`
	Document        string    `json:"document"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	Addresses       []Address `json:"addresses"`
	BirthDate       string    `json:"birth_date,omitempty"`
	Gender          string    `json:"gender,omitempty"`
	CompanyName     string    `json:"company_name,omitempty"`
	ResponsibleName string    `json:"responsible_name,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

func (p *Person) Normalize() {
	p.Kind = strings.ToUpper(strings.TrimSpace(p.Kind))
	p.Document = strings.TrimSpace(p.Document)
	p.Name = strings.TrimSpace(p.Name)
	p.Phone = strings.TrimSpace(p.Phone)
	p.Gender = strings.ToUpper(strings.TrimSpace(p.Gender))
	p.CompanyName = strings.TrimSpace(p.CompanyName)
	p.ResponsibleName = strings.TrimSpace(p.ResponsibleName)
	p.BirthDate = strings.TrimSpace(p.BirthDate)
	if p.Kind == KindPF {
		p.CompanyName = ""
		p.ResponsibleName = ""
	}
	if p.Kind == KindPJ {
		p.BirthDate = ""
		p.Gender = ""
	}
	out := make([]Address, 0, len(p.Addresses))
	for _, a := range p.Addresses {
		a.Normalize()
		if a.Empty() {
			continue
		}
		out = append(out, a)
	}
	p.Addresses = out
}

// Document (CPF/CNPJ) is intentionally not required — a quick counter-sale customer often has
// no document on hand at registration time.
func (p Person) Validate() error {
	if p.Kind != KindPF && p.Kind != KindPJ {
		return ErrInvalid
	}
	if p.Name == "" || p.Phone == "" {
		return ErrInvalid
	}
	return nil
}

type PersonRepository interface {
	Create(ctx context.Context, p Person) (Person, error)
	Update(ctx context.Context, p Person) error
	Get(ctx context.Context, id string) (Person, error)
	GetByDocument(ctx context.Context, document string) (Person, error)
}

type RoleRepository interface {
	Link(ctx context.Context, personID string) error
	Unlink(ctx context.Context, personID string) error
	Get(ctx context.Context, id string) (Person, error)
	List(ctx context.Context) ([]Person, error)
}

type PostalLookup interface {
	ByCEP(ctx context.Context, cep string) (Address, error)
	// SearchCEP is the reverse of ByCEP: finds the CEP(s) of a street. state (UF), city and
	// street (3+ chars) are required by the provider; district only narrows the result.
	SearchCEP(ctx context.Context, state, city, street, district string) ([]Address, error)
	ByGeo(ctx context.Context, lat, lng float64) (Address, error)
	Search(ctx context.Context, query string) (Address, error)
	SearchPlace(ctx context.Context, street, number, district, city, state, zip string) (Address, error)
}
