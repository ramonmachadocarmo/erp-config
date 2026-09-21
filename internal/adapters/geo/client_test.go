package geo

import (
	"context"
	"testing"

	"erp/services/config-service/internal/domain"
)

func TestFilterDistrict(t *testing.T) {
	list := []domain.Address{
		{Zip: "01310-100", District: "Bela Vista"},
		{Zip: "01311-000", District: "Cerqueira César"},
	}
	if got := filterDistrict(list, "cerqueira cesar"); len(got) != 1 || got[0].Zip != "01311-000" {
		t.Fatalf("accent-insensitive filter: %+v", got)
	}
	if got := filterDistrict(list, "inexistente"); len(got) != 2 {
		t.Fatalf("no match must keep the full list: %+v", got)
	}
	if got := filterDistrict(list, ""); len(got) != 2 {
		t.Fatalf("empty district must keep the full list: %+v", got)
	}
}

func TestSearchCEPValidatesInput(t *testing.T) {
	c := New()
	for _, tc := range [][3]string{{"", "São Paulo", "Paulista"}, {"SP", "Sã", "Paulista"}, {"SP", "São Paulo", "Pa"}} {
		if _, err := c.SearchCEP(context.Background(), tc[0], tc[1], tc[2], ""); err != domain.ErrInvalid {
			t.Fatalf("%v: %v", tc, err)
		}
	}
}
