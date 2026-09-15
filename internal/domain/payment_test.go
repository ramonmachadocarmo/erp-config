package domain

import "testing"

func TestPaymentTermValidate(t *testing.T) {
	ok := PaymentTerm{Code: "2X", Name: "2x", Installments: []InstallmentSpec{{Days: 30, Percent: 50}, {Days: 60, Percent: 50}}}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	vista := PaymentTerm{Code: "AVISTA", Name: "À vista", Installments: []InstallmentSpec{{Days: 0, Percent: 100}}}
	if err := vista.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (PaymentTerm{Code: "X", Name: "x"}).Validate(); err != ErrInvalid {
		t.Fatalf("empty: %v", err)
	}
	badSum := PaymentTerm{Code: "X", Name: "x", Installments: []InstallmentSpec{{Days: 0, Percent: 40}, {Days: 30, Percent: 40}}}
	if err := badSum.Validate(); err != ErrInvalid {
		t.Fatalf("sum: %v", err)
	}
	neg := PaymentTerm{Code: "X", Name: "x", Installments: []InstallmentSpec{{Days: -1, Percent: 100}}}
	if err := neg.Validate(); err != ErrInvalid {
		t.Fatalf("days: %v", err)
	}
}
