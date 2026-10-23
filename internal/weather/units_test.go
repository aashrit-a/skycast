package weather

import (
	"math"
	"testing"
)

func TestCompass(t *testing.T) {
	tests := map[int]string{0: "N", 22: "N", 23: "NE", 44: "NE", 45: "NE", 90: "E", 180: "S", 270: "W", 315: "NW", 359: "N", 360: "N", -90: "W"}
	for deg, want := range tests {
		if got := Compass(deg); got != want {
			t.Errorf("Compass(%d) = %s, want %s", deg, got, want)
		}
	}
}

func TestWindStrength(t *testing.T) {
	tests := []struct {
		speed float64
		u     Units
		want  float64
	}{
		{0, UnitsMetric, 0},
		{30, UnitsMetric, 0.5},
		{60, UnitsMetric, 1},
		{200, UnitsMetric, 1},
		{-5, UnitsMetric, 0},
		{37.282, UnitsImperial, 1},
	}
	for _, tt := range tests {
		if got := WindStrength(tt.speed, tt.u); math.Abs(got-tt.want) > 0.001 {
			t.Errorf("WindStrength(%v,%v) = %v, want %v", tt.speed, tt.u, got, tt.want)
		}
	}
}

func TestWindDrift(t *testing.T) {
	if d := WindDrift(60, 270, UnitsMetric); math.Abs(d-1) > 1e-9 {
		t.Errorf("westerly should drift right, got %v", d)
	}
	if d := WindDrift(60, 90, UnitsMetric); math.Abs(d+1) > 1e-9 {
		t.Errorf("easterly should drift left, got %v", d)
	}
	if d := WindDrift(60, 0, UnitsMetric); math.Abs(d) > 1e-9 {
		t.Errorf("northerly should not drift, got %v", d)
	}
}

func TestDefaultUnits(t *testing.T) {
	for cc, want := range map[string]Units{"US": UnitsImperial, "us": UnitsImperial, "LR": UnitsImperial, "MM": UnitsImperial, "JP": UnitsMetric, "": UnitsMetric} {
		if got := DefaultUnits(cc); got != want {
			t.Errorf("DefaultUnits(%q) = %v", cc, got)
		}
	}
}
