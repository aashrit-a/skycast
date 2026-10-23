package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/YOUR_USERNAME/skycast/internal/config"
	"github.com/YOUR_USERNAME/skycast/internal/scene"
	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

// Model is the Bubble Tea model.
type Model struct {
	cfg     config.Config
	scene   scene.Scene
	canvas  *scene.Canvas
	palette scene.Palette
	rng     *rand.Rand
	width   int
	height  int
	frame   int
}

// New builds the model. The palette must be built before the program starts.
func New(cfg config.Config, palette scene.Palette) Model {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	p := scene.Params{Condition: weather.Rain, Intensity: weather.Moderate, Wind: 0.3, IsDay: true, ASCIIOnly: cfg.ASCII}
	if cond, in, day, ok := scene.ByName(cfg.Scene); ok {
		p.Condition, p.Intensity, p.IsDay = cond, in, day
		if fi, ok := scene.ParseIntensity(cfg.Intensity); ok {
			p.Intensity = fi
		}
	}
	return Model{
		cfg:     cfg,
		scene:   scene.New(p, rng),
		canvas:  scene.NewCanvas(0, 0),
		palette: palette,
		rng:     rng,
	}
}

func (m Model) frameDuration() time.Duration {
	return time.Second / time.Duration(max(m.cfg.FPS, 1))
}

// Init starts the animation.
func (m Model) Init() tea.Cmd { return tick(m.frameDuration()) }

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.canvas.Resize(m.width, m.height)
		m.scene.Resize(m.width, m.height)
		for i := 0; i < 40; i++ {
			m.scene.Step()
		}
	case tickMsg:
		m.scene.Step()
		m.frame++
		return m, tick(m.frameDuration())
	}
	return m, nil
}

// View renders the current frame.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	m.canvas.Clear()
	m.scene.Draw(m.canvas)
	return m.canvas.Render(m.palette)
}
