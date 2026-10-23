package scene

import (
	"math"
	"math/rand"

	"github.com/aashrit-a/skycast/internal/weather"
)

// Particle is a moving glyph (raindrop, flake, hailstone, streak).
type Particle struct {
	X, Y   float64
	VX, VY float64
	Glyph  rune
	Color  ColorID
	Age    int
	Phase  float64 // for sinusoidal sway / twinkle
}

// spawnTop returns a particle at a random column just above row top.
func spawnTop(rng *rand.Rand, w int, top float64) Particle {
	return Particle{
		X:     rng.Float64() * float64(max(w, 1)),
		Y:     top - rng.Float64(),
		Phase: rng.Float64() * 2 * math.Pi,
	}
}

// wrapX wraps x into [0, w) so wind never empties the screen.
func wrapX(x float64, w int) float64 {
	fw := float64(max(w, 1))
	x = math.Mod(x, fw)
	if x < 0 {
		x += fw
	}
	return x
}

// respawnIfOut wraps p horizontally and reports whether it fell below
// bottom, in which case the caller should respawn it.
func respawnIfOut(p *Particle, w int, bottom float64) bool {
	p.X = wrapX(p.X, w)
	return p.Y >= bottom
}

// density returns how many particles to keep for an area and per-cell factor.
func density(area int, factor float64) int {
	if area <= 0 {
		return 0
	}
	return max(1, int(math.Round(float64(area)*factor)))
}

// byIntensity picks one of three values by intensity.
func byIntensity[T any](in weather.Intensity, light, moderate, heavy T) T {
	switch in {
	case weather.Moderate:
		return moderate
	case weather.Heavy:
		return heavy
	}
	return light
}

// between returns a uniform random value in [lo, hi).
func between(rng *rand.Rand, lo, hi float64) float64 {
	return lo + rng.Float64()*(hi-lo)
}

func clampF(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

func clampI(v, lo, hi int) int { return max(lo, min(hi, v)) }

// ix rounds a float coordinate to the nearest cell.
func ix(v float64) int { return int(math.Floor(v + 0.5)) }
