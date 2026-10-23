package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

func TestPanelASCII(t *testing.T) {
	st := newStyles(true, true)
	p := renderPanel(sampleReport(weather.UnitsMetric), st, panelWidth, fullPanelHeight, fixedNow)
	for _, want := range []string{"hum 82%", "wind 12 km/h NE", "rise 05:58", "set 17:02"} {
		if !strings.Contains(p, want) {
			t.Errorf("panel missing %q:\n%s", want, p)
		}
	}
	for _, r := range p {
		if r > 127 {
			t.Errorf("non-ASCII rune %q in ASCII panel", r)
		}
	}
}

func TestPanelTruncatesLongNames(t *testing.T) {
	r := sampleReport(weather.UnitsMetric)
	r.Location = "Llanfairpwllgwyngyllgogerychwyrndrobwllllantysiliogogogoch, United Kingdom"
	p := renderPanel(r, newStyles(true, false), panelWidth, fullPanelHeight, fixedNow)
	for _, l := range strings.Split(p, "\n") {
		if lipgloss.Width(l) != panelWidth {
			t.Errorf("line width %d: %q", lipgloss.Width(l), l)
		}
	}
	if !strings.Contains(p, "…") {
		t.Error("name not truncated")
	}
}

func TestPanelShortDaily(t *testing.T) {
	r := sampleReport(weather.UnitsMetric)
	r.Daily = nil
	_ = renderPanel(r, newStyles(true, false), panelWidth, fullPanelHeight, fixedNow)
	_ = renderWidePanel(r, newStyles(true, false), 50, stackedPanelH, fixedNow)
	r.Daily = sampleReport(weather.UnitsMetric).Daily[:2]
	if p := renderPanel(r, newStyles(true, false), panelWidth, fullPanelHeight, fixedNow); !strings.Contains(p, "Fri") {
		t.Error("partial forecast missing")
	}
}

func TestAgo(t *testing.T) {
	for d, want := range map[time.Duration]string{0: "just now", 5 * time.Minute: "5m ago", 2 * time.Hour: "2h ago", 72 * time.Hour: "3d ago"} {
		if got := ago(fixedNow.Add(-d), fixedNow); got != want {
			t.Errorf("ago(%v) = %q, want %q", d, got, want)
		}
	}
}
