package domain

import "testing"

func TestDistributionCenterValidate(t *testing.T) {
	ok := DistributionCenter{Name: "CD", WarehouseID: "w1", Lat: -3.1, Lng: -60.0}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (DistributionCenter{Name: "CD", WarehouseID: "w1"}).Validate(); err != ErrInvalid {
		t.Fatalf("%v", err)
	}
}

func TestDeliveryVehicleValidate(t *testing.T) {
	ok := DeliveryVehicle{Name: "Van", CapacityKg: 800, CapacityM3: 4}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (DeliveryVehicle{Name: "Van", CapacityKg: 800}).Validate(); err != ErrInvalid {
		t.Fatalf("%v", err)
	}
}
