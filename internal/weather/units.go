package weather

import (
	"math"
	"strings"
)

// TempSuffix returns the temperature unit label.
func TempSuffix(u Units) string {
	if u == UnitsImperial {
		return "°F"
	}
	return "°C"
}

// WindSuffix returns the wind speed unit label.
func WindSuffix(u Units) string {
	if u == UnitsImperial {
		return "mph"
	}
	return "km/h"
}

var compassPoints = [8]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}

// Compass converts a direction in degrees to one of 8 compass points.
func Compass(deg int) string {
	d := math.Mod(float64(deg), 360)
	if d < 0 {
		d += 360
	}
	return compassPoints[int(math.Floor((d+22.5)/45))%8]
}

// WindStrength normalizes a wind speed to 0..1, where 1 means 60 km/h or more.
func WindStrength(speed float64, u Units) float64 {
	kmh := speed
	if u == UnitsImperial {
		kmh = speed * 1.609344
	}
	return clamp01(kmh / 60)
}

// WindDrift returns the signed horizontal drift (-1..1) for scenes: positive
// means particles move right. Meteorological direction is where the wind
// comes FROM, so a westerly (270°) pushes things to the east (right).
func WindDrift(speed float64, dirDeg int, u Units) float64 {
	s := WindStrength(speed, u)
	v := -math.Sin(float64(dirDeg)*math.Pi/180) * s
	return math.Max(-1, math.Min(1, v))
}

// DefaultUnits picks imperial for the few countries that use it, else metric.
func DefaultUnits(countryCode string) Units {
	switch strings.ToUpper(countryCode) {
	case "US", "LR", "MM":
		return UnitsImperial
	}
	return UnitsMetric
}

// ParseUnits converts "metric"/"imperial" to Units; ok is false otherwise.
func ParseUnits(s string) (u Units, ok bool) {
	switch strings.ToLower(s) {
	case "metric":
		return UnitsMetric, true
	case "imperial":
		return UnitsImperial, true
	}
	return UnitsMetric, false
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
