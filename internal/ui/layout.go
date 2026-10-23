package ui

// LayoutMode selects how the scene and panel share the terminal.
type LayoutMode int

// Layout modes.
const (
	SideBySide LayoutMode = iota // scene left, panel right
	Stacked                      // scene on top, wide panel below
	Minimal                      // too small for a scene: one line of text
	Ambient                      // scene fills the whole terminal, no footer
	SceneOnly                    // panel hidden: scene plus footer
)

// Layout constants.
const (
	panelWidth      = 30
	panelGap        = 2
	maxSceneHeight  = 30
	stackedPanelH   = 9
	minSceneW       = 10
	minSceneH       = 4
	minSideBySideH  = 9 // compact panel (no forecast) needs this many rows
	fullPanelHeight = 17
)

// Layout is the computed placement of the scene and panel.
type Layout struct {
	Mode           LayoutMode
	SceneW, SceneH int
	PanelW, PanelH int
}

// HasFooter reports whether a footer row is shown.
func (l Layout) HasFooter() bool { return l.Mode != Ambient }

// ComputeLayout decides the layout for a terminal size. One row is
// reserved for the footer unless ambient.
func ComputeLayout(termW, termH int, showPanel, ambient bool) Layout {
	if ambient {
		if termW < minSceneW || termH < minSceneH {
			return Layout{Mode: Minimal}
		}
		return Layout{Mode: Ambient, SceneW: termW, SceneH: termH}
	}
	avail := termH - 1
	if !showPanel {
		// NOTE: "panel hidden" keeps the footer, so it's its own mode
		// rather than Ambient.
		if termW < minSceneW || avail < minSceneH {
			return Layout{Mode: Minimal}
		}
		return Layout{Mode: SceneOnly, SceneW: termW, SceneH: avail}
	}
	switch {
	case termW >= 72 && avail >= minSideBySideH:
		return Layout{
			Mode:   SideBySide,
			SceneW: termW - panelWidth - panelGap,
			SceneH: min(avail, maxSceneHeight),
			PanelW: panelWidth,
			PanelH: min(avail, fullPanelHeight),
		}
	case termW >= 40 && termH >= 22:
		return Layout{
			Mode:   Stacked,
			SceneW: termW,
			SceneH: min(avail-stackedPanelH, maxSceneHeight),
			PanelW: termW,
			PanelH: stackedPanelH,
		}
	}
	return Layout{Mode: Minimal}
}
