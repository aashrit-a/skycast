// Package scene implements the animated ASCII weather scenes.
package scene

import (
	"math"
	"math/rand"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

// Params selects and configures a scene.
type Params struct {
	Condition  weather.Condition
	Intensity  weather.Intensity
	Wind       float64 // signed drift -1..1 from weather.WindDrift: negative = blowing left
	IsDay      bool
	ASCIIOnly  bool // --ascii: no Unicode glyphs
	SnowGrains bool // WMO 77: small fast grains instead of flakes
}

// Scene is an animated weather scene.
type Scene interface {
	// Resize is called on start and whenever the scene area changes size.
	Resize(w, h int)
	// Step advances the simulation by one frame.
	Step()
	// Draw paints the current frame into the canvas (canvas already sized w×h and cleared).
	Draw(c *Canvas)
	// Name returns a short identifier, e.g. "rain".
	Name() string
}

// New returns the scene for the given params. rng is injected for
// deterministic tests. Strong wind adds a streak overlay automatically.
func New(p Params, rng *rand.Rand) Scene {
	p.Wind = clampF(p.Wind, -1, 1)
	var s Scene
	switch p.Condition {
	case weather.ClearDay:
		s = newSun(p, rng)
	case weather.ClearNight:
		s = newNight(p, rng)
	case weather.PartlyCloudy, weather.Overcast:
		s = newClouds(p, rng)
	case weather.Fog:
		s = newFog(p, rng)
	case weather.Snow:
		s = newSnow(p, rng)
	case weather.Thunderstorm, weather.Hail:
		s = newStorm(p, rng)
	default:
		s = newRain(p, rng)
	}
	if math.Abs(p.Wind) > windOverlayThreshold {
		s = WithWind(s, p.Wind, rng)
	}
	// Give every scene a valid (tiny) size so Step/Draw are safe even
	// before the first real Resize.
	s.Resize(1, 1)
	return s
}

// unwrap strips decorators (such as the wind overlay) to reach the base scene.
func unwrap(s Scene) Scene {
	for {
		u, ok := s.(interface{ base() Scene })
		if !ok {
			return s
		}
		s = u.base()
	}
}

// ByName maps a --scene name to its condition, default intensity and
// whether it is a daytime scene. ok is false for unknown names.
func ByName(name string) (cond weather.Condition, in weather.Intensity, isDay, ok bool) {
	switch name {
	case "clear":
		return weather.ClearDay, weather.Light, true, true
	case "night":
		return weather.ClearNight, weather.Light, false, true
	case "cloudy":
		return weather.PartlyCloudy, weather.Moderate, true, true
	case "overcast":
		return weather.Overcast, weather.Heavy, true, true
	case "fog":
		return weather.Fog, weather.Moderate, true, true
	case "drizzle":
		return weather.Drizzle, weather.Moderate, true, true
	case "rain":
		return weather.Rain, weather.Moderate, true, true
	case "freezing":
		return weather.FreezingRain, weather.Moderate, true, true
	case "snow":
		return weather.Snow, weather.Moderate, true, true
	case "storm":
		return weather.Thunderstorm, weather.Moderate, true, true
	case "hail":
		return weather.Hail, weather.Moderate, true, true
	}
	return 0, 0, false, false
}

// ParseIntensity converts "light"/"moderate"/"heavy" to an Intensity.
func ParseIntensity(s string) (weather.Intensity, bool) {
	switch s {
	case "light":
		return weather.Light, true
	case "moderate":
		return weather.Moderate, true
	case "heavy":
		return weather.Heavy, true
	}
	return weather.Light, false
}
