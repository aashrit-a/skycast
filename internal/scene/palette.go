package scene

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorID indexes a color in the active Palette. ColNone means "terminal default".
type ColorID uint8

// Palette color slots.
const (
	ColNone ColorID = iota
	ColSun
	ColSunGlow
	ColSunDim
	ColMoon
	ColStar
	ColStarDim
	ColCloudLight
	ColCloudDark
	ColRain
	ColRainLight
	ColSplash
	ColIce
	ColSnow
	ColSnowDim
	ColBolt
	ColFog1
	ColFog2
	ColFog3
	ColWind
	ColGround
	numColors
)

// colors maps each slot to a light-background / dark-background pair.
var colors = [numColors]lipgloss.AdaptiveColor{
	ColSun:        {Light: "#D78700", Dark: "#FFD75F"},
	ColSunGlow:    {Light: "#AF5F00", Dark: "#FFAF00"},
	ColSunDim:     {Light: "#AF8700", Dark: "#AF8700"},
	ColMoon:       {Light: "#5F5F87", Dark: "#EEEEDD"},
	ColStar:       {Light: "#303030", Dark: "#FFFFFF"},
	ColStarDim:    {Light: "#9E9E9E", Dark: "#6C6C6C"},
	ColCloudLight: {Light: "#808080", Dark: "#D0D0D0"},
	ColCloudDark:  {Light: "#4E4E4E", Dark: "#808080"},
	ColRain:       {Light: "#005FAF", Dark: "#5FAFFF"},
	ColRainLight:  {Light: "#5F87AF", Dark: "#3A6E9F"},
	ColSplash:     {Light: "#0087AF", Dark: "#87D7FF"},
	ColIce:        {Light: "#00AFAF", Dark: "#AFFFFF"},
	ColSnow:       {Light: "#6C6C6C", Dark: "#FFFFFF"},
	ColSnowDim:    {Light: "#B2B2B2", Dark: "#A8A8A8"},
	ColBolt:       {Light: "#AF8700", Dark: "#FFFF87"},
	ColFog1:       {Light: "#8A8A8A", Dark: "#C6C6C6"},
	ColFog2:       {Light: "#A8A8A8", Dark: "#9E9E9E"},
	ColFog3:       {Light: "#C6C6C6", Dark: "#6C6C6C"},
	ColWind:       {Light: "#87879F", Dark: "#8787AF"},
	ColGround:     {Light: "#5F875F", Dark: "#5F875F"},
}

// Palette holds pre-rendered escape sequences for every color slot.
//
// NOTE: lipgloss.Style.Render is too slow to call for every run at 12+ fps,
// so each style is rendered once around a marker and split into an ANSI
// prefix and suffix. Rendering a run is then plain string concatenation.
type Palette struct {
	prefix, suffix [numColors * 2]string // index: color*2 + bold
}

// NewPalette builds the palette. With noColor every style is unstyled.
// Call it once at startup (before Bubble Tea takes over the terminal),
// since resolving adaptive colors may query the terminal background.
func NewPalette(noColor bool) Palette {
	var p Palette
	if noColor {
		return p
	}
	for id := ColorID(1); id < numColors; id++ {
		base := lipgloss.NewStyle().Foreground(colors[id])
		p.set(id, false, base)
		p.set(id, true, base.Bold(true))
	}
	// Bold with no color still matters (e.g. lightning on the default color).
	p.set(ColNone, true, lipgloss.NewStyle().Bold(true))
	return p
}

const marker = "\x00"

func (p *Palette) set(id ColorID, bold bool, s lipgloss.Style) {
	out := s.Render(marker)
	pre, suf, ok := strings.Cut(out, marker)
	if !ok {
		return
	}
	i := int(id)*2 + b2i(bold)
	p.prefix[i], p.suffix[i] = pre, suf
}

func (p *Palette) codes(id ColorID, bold bool) (string, string) {
	if id >= numColors {
		id = ColNone
	}
	i := int(id)*2 + b2i(bold)
	return p.prefix[i], p.suffix[i]
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
