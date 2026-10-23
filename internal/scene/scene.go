// Package scene implements the animated ASCII weather scenes.
package scene

import (
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
	case weather.Drizzle, weather.Rain, weather.FreezingRain:
		s = newRain(p, rng)
	default:
		// NOTE: remaining scenes arrive in phase 4; rain stands in.
		s = newRain(p, rng)
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
