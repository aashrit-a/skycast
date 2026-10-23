package weather

import "fmt"

// Classify maps a WMO weather code to a condition, intensity and label.
func Classify(code int, isDay bool) (Condition, Intensity, string) {
	clear := ClearNight
	if isDay {
		clear = ClearDay
	}
	switch code {
	case 0:
		return clear, Light, "Clear sky"
	case 1:
		return clear, Light, "Mainly clear"
	case 2:
		return PartlyCloudy, Moderate, "Partly cloudy"
	case 3:
		return Overcast, Heavy, "Overcast"
	case 45:
		return Fog, Moderate, "Fog"
	case 48:
		return Fog, Heavy, "Rime fog"
	case 51:
		return Drizzle, Light, "Light drizzle"
	case 53:
		return Drizzle, Moderate, "Moderate drizzle"
	case 55:
		return Drizzle, Heavy, "Dense drizzle"
	case 56:
		return FreezingRain, Light, "Light freezing drizzle"
	case 57:
		return FreezingRain, Heavy, "Dense freezing drizzle"
	case 61:
		return Rain, Light, "Slight rain"
	case 63:
		return Rain, Moderate, "Moderate rain"
	case 65:
		return Rain, Heavy, "Heavy rain"
	case 66:
		return FreezingRain, Light, "Light freezing rain"
	case 67:
		return FreezingRain, Heavy, "Heavy freezing rain"
	case 71:
		return Snow, Light, "Slight snow"
	case 73:
		return Snow, Moderate, "Moderate snow"
	case 75:
		return Snow, Heavy, "Heavy snow"
	case 77:
		return Snow, Light, "Snow grains"
	case 80:
		return Rain, Light, "Rain showers"
	case 81:
		return Rain, Moderate, "Rain showers"
	case 82:
		return Rain, Heavy, "Violent rain showers"
	case 85:
		return Snow, Moderate, "Snow showers"
	case 86:
		return Snow, Heavy, "Heavy snow showers"
	case 95:
		return Thunderstorm, Moderate, "Thunderstorm"
	case 96:
		return Hail, Moderate, "Thunderstorm with hail"
	case 99:
		return Hail, Heavy, "Thunderstorm with heavy hail"
	}
	return Overcast, Moderate, fmt.Sprintf("Unknown (%d)", code)
}

// IsSnowGrains reports whether code is WMO 77 (snow grains), which the snow
// scene renders differently.
func IsSnowGrains(code int) bool { return code == 77 }

// Small 3-line, 7-column icons for the forecast row.
var (
	iconSun    = []string{" \\ | / ", "-- O --", " / | \\ "}
	iconMoon   = []string{"  .-.  ", " (  (  ", "  '-'  "}
	iconPartly = []string{"  \\ .-.", " -(   )", "  '--' "}
	iconCloud  = []string{"  .--. ", " (    )", "(_____)"}
	iconRain   = []string{"  .--. ", " (____)", " ' ' ' "}
	iconSnow   = []string{"  .--. ", " (____)", " * * * "}
	iconStorm  = []string{"  .--. ", " (____)", " ' _/ '"}
	iconFog    = []string{"_ - _ -", " _ - _ ", "_ - _ -"}
)

// SmallIcon returns a static 3-line, 7-column ASCII icon for a weather code.
func SmallIcon(code int, isDay bool) []string {
	cond, _, _ := Classify(code, isDay)
	var icon []string
	switch cond {
	case ClearDay:
		icon = iconSun
	case ClearNight:
		icon = iconMoon
	case PartlyCloudy:
		icon = iconPartly
	case Fog:
		icon = iconFog
	case Drizzle, Rain, FreezingRain:
		icon = iconRain
	case Snow:
		icon = iconSnow
	case Thunderstorm, Hail:
		icon = iconStorm
	default:
		icon = iconCloud
	}
	out := make([]string, len(icon))
	copy(out, icon)
	return out
}
