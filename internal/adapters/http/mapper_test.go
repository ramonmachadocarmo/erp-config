package httpadapter

import (
	"testing"

	"erp-schema/model"
)

func TestToDomainAddress(t *testing.T) {
	lat, lng := -23.5, -46.6
	id := "addr-1"
	m := model.Address{
		ID:         &id,
		Alias:      "Home",
		Zip:        "01000-000",
		Street:     "Main St",
		Number:     "100",
		Complement: "Apt 1",
		District:   "Downtown",
		City:       "São Paulo",
		State:      "SP",
		Lat:        &lat,
		Lng:        &lng,
	}

	got := toDomainAddress(m)

	if got.ID != id {
		t.Errorf("ID = %q, want %q", got.ID, id)
	}
	if got.Alias != m.Alias || got.Zip != m.Zip || got.Street != m.Street ||
		got.Number != m.Number || got.Complement != m.Complement ||
		got.District != m.District || got.City != m.City || got.State != m.State {
		t.Errorf("mapped fields do not match source: got %+v, from %+v", got, m)
	}
	if got.Lat == nil || *got.Lat != lat {
		t.Errorf("Lat = %v, want %v", got.Lat, lat)
	}
	if got.Lng == nil || *got.Lng != lng {
		t.Errorf("Lng = %v, want %v", got.Lng, lng)
	}
}

func TestToDomainAddressOmitsUnsetIdentifierAndCoords(t *testing.T) {
	m := model.Address{
		Alias:    "Work",
		Zip:      "02000-000",
		Street:   "Second St",
		Number:   "200",
		District: "Uptown",
		City:     "Rio de Janeiro",
		State:    "RJ",
	}

	got := toDomainAddress(m)

	if got.ID != "" {
		t.Errorf("ID = %q, want empty", got.ID)
	}
	if got.Lat != nil {
		t.Errorf("Lat = %v, want nil", got.Lat)
	}
	if got.Lng != nil {
		t.Errorf("Lng = %v, want nil", got.Lng)
	}
}

func TestToDomainAddresses(t *testing.T) {
	ms := []model.Address{
		{Alias: "Home", Zip: "1", Street: "A", Number: "1", District: "D1", City: "C1", State: "S1"},
		{Alias: "Work", Zip: "2", Street: "B", Number: "2", District: "D2", City: "C2", State: "S2"},
	}

	got := toDomainAddresses(ms)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Alias != "Home" || got[1].Alias != "Work" {
		t.Errorf("addresses not mapped in order: %+v", got)
	}
}

func TestToDomainAddressesEmpty(t *testing.T) {
	got := toDomainAddresses(nil)
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}
