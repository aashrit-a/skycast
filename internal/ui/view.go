package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aashrit-a/skycast/internal/location"
)

var (
	spinnerUnicode = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerASCII   = []string{"|", "/", "-", "\\"}
)

func (m Model) spinnerFrame() string {
	frames := spinnerUnicode
	if m.cfg.ASCII {
		frames = spinnerASCII
	}
	return frames[m.spinner%len(frames)]
}

// View renders the whole screen; output is exactly m.height lines.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	switch m.state {
	case stateLoading:
		msg := "fetching weather for " + placeLabel(m.cfg) + m.styles.ell + " " + m.spinnerFrame()
		return m.place(m.width, m.height, m.styles.trunc(msg, m.width))
	case stateError:
		return m.place(m.width, m.height, m.errorView())
	}

	l := m.layout
	var body string
	switch l.Mode {
	case Minimal:
		body = m.styles.trunc(minimalLine(m.report, m.styles), m.width)
	case Ambient, SceneOnly:
		body = m.renderScene()
	case SideBySide:
		panel := renderPanel(m.report, m.styles, l.PanelW, l.PanelH, m.now())
		body = lipgloss.JoinHorizontal(lipgloss.Center, m.renderScene(), strings.Repeat(" ", panelGap), panel)
	case Stacked:
		panel := renderWidePanel(m.report, m.styles, l.PanelW, l.PanelH, m.now())
		body = lipgloss.JoinVertical(lipgloss.Left, m.renderScene(), panel)
	}
	if !l.HasFooter() || m.height < 2 {
		return m.place(m.width, m.height, body)
	}
	return m.place(m.width, m.height-1, body) + "\n" + m.footer()
}

func (m Model) renderScene() string {
	m.canvas.Clear()
	if m.scene != nil {
		m.scene.Draw(m.canvas)
	}
	return m.canvas.Render(m.palette)
}

// place centers content in a w×h box, clipping anything that overflows.
func (m Model) place(w, h int, content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func (m Model) footer() string {
	st := m.styles
	keys := footerKeys
	if m.demo != nil {
		keys = footerDemoKeys
		if st.ascii {
			keys = append([]string{"</> scene"}, keys[1:]...)
		}
	}
	text := "? help"
	if m.showHelp {
		text = st.join(keys...)
	}
	if m.demo != nil {
		text = st.join(m.demo.label(st), text)
	}
	if m.fetching && m.state == stateReady {
		text = st.join(text, "updating "+m.spinnerFrame())
	}
	return st.dim.Render(st.trunc(text, m.width))
}

func (m Model) errorView() string {
	hint := m.styles.join("r retry", "q quit")
	if errors.Is(m.err, location.ErrLocationNotFound) {
		hint = `check the spelling, or add a country: skycast "paris, fr"` + "\n" + hint
	}
	w := max(10, m.width-4)
	msg := lipgloss.NewStyle().Width(min(w, 60)).Align(lipgloss.Center).Render(m.err.Error())
	return lipgloss.JoinVertical(lipgloss.Center,
		m.styles.err.Render(msg),
		"",
		m.styles.dim.Render(m.truncateLines(hint, m.width)),
	)
}

func (m Model) truncateLines(s string, w int) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = m.styles.trunc(l, w)
	}
	return strings.Join(lines, "\n")
}
