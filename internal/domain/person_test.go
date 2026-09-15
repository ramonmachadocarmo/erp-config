package domain

import "testing"

func TestPersonValidateRequired(t *testing.T) {
	p := Person{Kind: KindPF, Document: "11111111111", Name: "TESTE", Phone: "11111111"}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	// Document is optional — e.g. a quick counter-sale customer with no CPF on hand.
	p.Document = ""
	if err := p.Validate(); err != nil {
		t.Fatalf("document should be optional: %v", err)
	}
	p.Name = ""
	if err := p.Validate(); err != ErrInvalid {
		t.Fatalf("name: %v", err)
	}
	p.Name = "TESTE"
	p.Phone = ""
	if err := p.Validate(); err != ErrInvalid {
		t.Fatalf("phone: %v", err)
	}
}

func TestPersonNormalizeAddresses(t *testing.T) {
	p := Person{
		Kind: KindPF, Document: "1", Name: "A", Phone: "2",
		CompanyName: "x",
		Addresses: []Address{
			{Alias: " Casa ", Street: " Rua A ", City: " sp ", State: "sp"},
			{},
		},
	}
	p.Normalize()
	if p.CompanyName != "" || len(p.Addresses) != 1 || p.Addresses[0].Alias != "Casa" || p.Addresses[0].State != "SP" {
		t.Fatalf("%+v", p)
	}
}
