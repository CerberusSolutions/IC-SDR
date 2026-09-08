package screens

import "testing"

func TestCalibrateSMeterMatchesOriginalCurve(t *testing.T) {
	tests := []struct {
		spectrum float32
		want     float32
	}{
		{-20, -73},
		{-34, -101},
		{-60, -130},
		{40, -13},
	}
	for _, test := range tests {
		if got := calibrateSMeter(test.spectrum); got != test.want {
			t.Errorf("calibrateSMeter(%v) = %v, want %v", test.spectrum, got, test.want)
		}
	}
}

func TestSMeterLabels(t *testing.T) {
	tests := []struct {
		db   int
		want string
	}{
		{-122, "S0"},
		{-121, "S1"},
		{-109, "S3"},
		{-73, "S9"},
		{-63, "S9+10"},
		{-53, "S9+20"},
	}
	for _, test := range tests {
		if got := sMeterLabel(test.db); got != test.want {
			t.Errorf("sMeterLabel(%d) = %q, want %q", test.db, got, test.want)
		}
	}
}
