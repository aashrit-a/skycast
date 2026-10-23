package ui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

// asciiBorder replaces the rounded border in --ascii mode.
var asciiBorder = lipgloss.Border{
	Top: "-", Bottom: "-", Left: "|", Right: "|",
	TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
}

// styles are the text styles and glyphs used outside the scene.
type styles struct {
	border lipgloss.Style
	title  lipgloss.Style
	bold   lipgloss.Style
	dim    lipgloss.Style
	err    lipgloss.Style

	ascii bool
	sep   string // between items on one line
	ell   string // truncation marker
	deg   string // degree sign
}

func newStyles(noColor, ascii bool) styles {
	s := styles{
		border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
		title:  lipgloss.NewStyle().Bold(true),
		bold:   lipgloss.NewStyle().Bold(true),
		dim:    lipgloss.NewStyle(),
		err:    lipgloss.NewStyle().Bold(true),
		ascii:  ascii,
		sep:    " · ",
		ell:    "…",
		deg:    "°",
	}
	if ascii {
		s.border = s.border.Border(asciiBorder)
		s.sep, s.ell, s.deg = " | ", "...", ""
	}
	if noColor {
		return s
	}
	gray := lipgloss.AdaptiveColor{Light: "#8A8A8A", Dark: "#6C6C6C"}
	s.border = s.border.BorderForeground(gray)
	s.title = s.title.Foreground(lipgloss.AdaptiveColor{Light: "#005F87", Dark: "#87D7FF"})
	s.dim = s.dim.Foreground(gray)
	s.err = s.err.Foreground(lipgloss.AdaptiveColor{Light: "#AF0000", Dark: "#FF5F5F"})
	return s
}

// join joins items with the separator.
func (st styles) join(items ...string) string { return strings.Join(items, st.sep) }

// temp formats a temperature such as "18°C" ("18C" in ASCII mode).
func (st styles) temp(v float64, u weather.Units) string {
	return round(v) + strings.Replace(weather.TempSuffix(u), "°", st.deg, 1)
}

// trunc shortens s to at most w columns, ending with the ellipsis if cut.
func (st styles) trunc(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	r := []rune(s)
	ew := lipgloss.Width(st.ell)
	for len(r) > 0 && lipgloss.Width(string(r))+ew > w {
		r = r[:len(r)-1]
	}
	if ew > w {
		return string(r)
	}
	return string(r) + st.ell
}

// round formats a value as a whole number without "-0".
func round(v float64) string {
	return strconv.Itoa(int(math.Round(v)))
}

func formatCoords(lat, lon float64) string {
	return fmt.Sprintf("%.2f, %.2f", lat, lon)
}

// ago formats how long ago t was, relative to now.
func ago(t, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

// currentLines is the "now" block: place, condition, temps, details.
func currentLines(r *weather.Report, st styles, w int) []string {
	c := r.Current
	_, _, label := weather.Classify(c.Code, c.IsDay)
	wind := fmt.Sprintf("%s %s %s", round(c.WindSpeed), weather.WindSuffix(r.Units), weather.Compass(c.WindDirDeg))
	lines := []string{
		st.title.Render(st.trunc(r.Location, w)),
		st.trunc(label, w),
		"",
		st.bold.Render(st.temp(c.Temp, r.Units)) + "   feels " + st.temp(c.FeelsLike, r.Units),
	}
	// NOTE: 💨 instead of 🌬 — 🌬 has no default emoji presentation, so
	// terminals disagree on its width and the panel border would wobble.
	if st.ascii {
		lines = append(lines, st.trunc(fmt.Sprintf("hum %d%%  wind %s", c.Humidity, wind), w))
	} else {
		lines = append(lines, st.trunc(fmt.Sprintf("💧 %d%%   💨 %s", c.Humidity, wind), w))
	}
	if len(r.Daily) > 0 && !r.Daily[0].Sunrise.IsZero() {
		rise, set := r.Daily[0].Sunrise.Format("15:04"), r.Daily[0].Sunset.Format("15:04")
		if st.ascii {
			lines = append(lines, "rise "+rise+"  set "+set)
		} else {
			lines = append(lines, "☀ "+rise+"   ☾ "+set)
		}
	}
	return lines
}

const forecastCol = 8

// forecastLines renders up to n days (Daily[1..n]) as aligned columns.
func forecastLines(r *weather.Report, st styles, n int) []string {
	days := r.Daily
	if len(days) > 1 {
		days = days[1:]
	} else {
		days = nil
	}
	if len(days) > n {
		days = days[:n]
	}
	if len(days) == 0 {
		return nil
	}
	rows := make([][]string, 6)
	for _, d := range days {
		icon := weather.SmallIcon(d.Code, true)
		cells := []string{
			d.Date.Weekday().String()[:3],
			icon[0], icon[1], icon[2],
			round(d.High) + "/" + round(d.Low) + st.deg,
			strconv.Itoa(d.PrecipChance) + "%",
		}
		for i, c := range cells {
			rows[i] = append(rows[i], c)
		}
	}
	out := make([]string, len(rows))
	for i, row := range rows {
		var b strings.Builder
		for j, c := range row {
			if j < len(row)-1 {
				c = padRight(c, forecastCol)
			}
			b.WriteString(c)
		}
		s := strings.TrimRight(b.String(), " ")
		switch i {
		case 0:
			s = st.bold.Render(s)
		case 5:
			s = st.dim.Render(s)
		}
		out[i] = s
	}
	return out
}

func padRight(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

func staleLine(r *weather.Report, st styles, now time.Time) string {
	if !r.Stale {
		return ""
	}
	return st.dim.Render(st.join("offline", "updated "+ago(r.FetchedAt, now)))
}

// renderPanel draws the tall side panel: width and maxH include the border.
func renderPanel(r *weather.Report, st styles, width, maxH int, now time.Time) string {
	inner := width - 4 // border + padding
	lines := currentLines(r, st, inner)
	stale := staleLine(r, st, now)
	extra := 0
	if stale != "" {
		extra = 2
	}
	if fc := forecastLines(r, st, 3); fc != nil && len(lines)+1+len(fc)+extra+2 <= maxH {
		lines = append(append(lines, ""), fc...)
	}
	if stale != "" && len(lines)+extra+2 <= maxH {
		lines = append(lines, "", stale)
	}
	for len(lines)+2 > maxH && len(lines) > 1 {
		lines = lines[:len(lines)-1]
	}
	return st.border.Width(width - 2).Render(strings.Join(lines, "\n"))
}

// renderWidePanel draws the stacked layout's panel: current conditions on
// the left, as many forecast days as fit on the right.
func renderWidePanel(r *weather.Report, st styles, width, height int, now time.Time) string {
	inner := width - 4
	left := currentLines(r, st, min(inner, 28))
	if stale := staleLine(r, st, now); stale != "" {
		left = append(left, stale)
	}
	leftW := 0
	for _, l := range left {
		leftW = max(leftW, lipgloss.Width(l))
	}
	const gap = 3
	n := 0
	for n < 3 && leftW+gap+(n+1)*forecastCol-1 <= inner {
		n++
	}
	block := strings.Join(left, "\n")
	if fc := forecastLines(r, st, n); fc != nil {
		leftBlock := lipgloss.NewStyle().Width(leftW + gap).Render(block)
		block = lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, strings.Join(fc, "\n"))
	}
	lines := strings.Split(block, "\n")
	if len(lines) > height-2 {
		lines = lines[:max(0, height-2)]
	}
	return st.border.Width(width - 2).Height(height - 2).Render(strings.Join(lines, "\n"))
}

// minimalLine is the one-line summary for tiny terminals.
func minimalLine(r *weather.Report, st styles) string {
	c := r.Current
	_, _, label := weather.Classify(c.Code, c.IsDay)
	parts := []string{
		r.Location,
		st.temp(c.Temp, r.Units),
		label,
		round(c.WindSpeed) + " " + weather.WindSuffix(r.Units) + " " + weather.Compass(c.WindDirDeg),
	}
	if r.Stale {
		parts = append(parts, "offline")
	}
	return st.join(parts...)
}
