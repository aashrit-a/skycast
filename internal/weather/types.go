// Package weather fetches forecasts from Open-Meteo and classifies them
// into the conditions skycast knows how to animate.
package weather

import "time"

// Units selects the measurement system used for requests and display.
type Units int

// Supported unit systems.
const (
	UnitsMetric Units = iota
	UnitsImperial
)

// String returns "metric" or "imperial".
func (u Units) String() string {
	if u == UnitsImperial {
		return "imperial"
	}
	return "metric"
}

// Condition is the broad weather category that decides which scene is shown.
type Condition int

// Known conditions.
const (
	ClearDay Condition = iota
	ClearNight
	PartlyCloudy
	Overcast
	Fog
	Drizzle
	Rain
	FreezingRain
	Snow
	Thunderstorm
	Hail
)

var conditionNames = [...]string{
	"clear-day", "clear-night", "partly-cloudy", "overcast", "fog",
	"drizzle", "rain", "freezing-rain", "snow", "thunderstorm", "hail",
}

// String returns a short kebab-case name for the condition.
func (c Condition) String() string {
	if c < 0 || int(c) >= len(conditionNames) {
		return "unknown"
	}
	return conditionNames[c]
}

// Intensity is how strong a condition is (drizzle vs. downpour).
type Intensity int

// Intensity levels.
const (
	Light Intensity = iota
	Moderate
	Heavy
)

// String returns "light", "moderate" or "heavy".
func (i Intensity) String() string {
	switch i {
	case Moderate:
		return "moderate"
	case Heavy:
		return "heavy"
	default:
		return "light"
	}
}

// Current holds the present conditions at a location.
type Current struct {
	Temp, FeelsLike float64
	Humidity        int
	WindSpeed       float64
	WindDirDeg      int
	Precip          float64
	IsDay           bool
	Code            int
}

// DailyForecast is one day of the forecast.
type DailyForecast struct {
	Date            time.Time
	Code            int
	High, Low       float64
	PrecipChance    int
	Sunrise, Sunset time.Time
}

// Report is everything skycast displays for one location.
type Report struct {
	Location  string // "Tokyo, Japan"
	Units     Units
	Current   Current
	Daily     []DailyForecast // up to 4 entries, [0] = today
	FetchedAt time.Time
	Stale     bool // served from an old cache because the network failed
}
