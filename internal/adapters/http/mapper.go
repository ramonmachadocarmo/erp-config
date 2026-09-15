package httpadapter

import (
	"erp-schema/model"
	"erp/services/config-service/internal/domain"
)

func toDomainAddress(m model.Address) domain.Address {
	addr := domain.Address{
		Alias:      m.Alias,
		Zip:        m.Zip,
		Street:     m.Street,
		Number:     m.Number,
		Complement: m.Complement,
		District:   m.District,
		City:       m.City,
		State:      m.State,
		Lat:        m.Lat,
		Lng:        m.Lng,
	}
	if m.ID != nil {
		addr.ID = *m.ID
	}
	return addr
}

func toDomainAddresses(ms []model.Address) []domain.Address {
	out := make([]domain.Address, len(ms))
	for i, m := range ms {
		out[i] = toDomainAddress(m)
	}
	return out
}
