package weather

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		code  int
		isDay bool
		cond  Condition
		in    Intensity
		label string
	}{
		{0, true, ClearDay, Light, "Clear sky"},
		{0, false, ClearNight, Light, "Clear sky"},
		{1, true, ClearDay, Light, "Mainly clear"},
		{1, false, ClearNight, Light, "Mainly clear"},
		{2, true, PartlyCloudy, Moderate, "Partly cloudy"},
		{3, true, Overcast, Heavy, "Overcast"},
		{45, true, Fog, Moderate, "Fog"},
		{48, true, Fog, Heavy, "Rime fog"},
		{51, true, Drizzle, Light, "Light drizzle"},
		{53, true, Drizzle, Moderate, "Moderate drizzle"},
		{55, true, Drizzle, Heavy, "Dense drizzle"},
		{56, true, FreezingRain, Light, "Light freezing drizzle"},
		{57, true, FreezingRain, Heavy, "Dense freezing drizzle"},
		{61, true, Rain, Light, "Slight rain"},
		{63, true, Rain, Moderate, "Moderate rain"},
		{65, true, Rain, Heavy, "Heavy rain"},
		{66, true, FreezingRain, Light, "Light freezing rain"},
		{67, true, FreezingRain, Heavy, "Heavy freezing rain"},
		{71, true, Snow, Light, "Slight snow"},
		{73, true, Snow, Moderate, "Moderate snow"},
		{75, true, Snow, Heavy, "Heavy snow"},
		{77, true, Snow, Light, "Snow grains"},
		{80, true, Rain, Light, "Rain showers"},
		{81, true, Rain, Moderate, "Rain showers"},
		{82, true, Rain, Heavy, "Violent rain showers"},
		{85, true, Snow, Moderate, "Snow showers"},
		{86, true, Snow, Heavy, "Heavy snow showers"},
		{95, true, Thunderstorm, Moderate, "Thunderstorm"},
		{96, true, Hail, Moderate, "Thunderstorm with hail"},
		{99, true, Hail, Heavy, "Thunderstorm with heavy hail"},
		{42, true, Overcast, Moderate, "Unknown (42)"},
	}
	for _, tt := range tests {
		c, i, l := Classify(tt.code, tt.isDay)
		if c != tt.cond || i != tt.in || l != tt.label {
			t.Errorf("Classify(%d,%v) = %v,%v,%q; want %v,%v,%q", tt.code, tt.isDay, c, i, l, tt.cond, tt.in, tt.label)
		}
	}
}

func TestSmallIconShape(t *testing.T) {
	for _, code := range []int{0, 2, 3, 45, 61, 71, 95, 99, 1234} {
		for _, day := range []bool{true, false} {
			icon := SmallIcon(code, day)
			if len(icon) != 3 {
				t.Fatalf("code %d: %d lines", code, len(icon))
			}
			for _, line := range icon {
				if len([]rune(line)) != 7 {
					t.Errorf("code %d: line %q is not 7 columns", code, line)
				}
			}
		}
	}
}
