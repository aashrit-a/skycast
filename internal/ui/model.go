package ui

import (
	"context"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/YOUR_USERNAME/skycast/internal/config"
	"github.com/YOUR_USERNAME/skycast/internal/scene"
	"github.com/YOUR_USERNAME/skycast/internal/weather"
)

type state int

const (
	stateLoading state = iota
	stateReady
	stateError
)

// prewarmSteps fills the scene with particles before the first frame.
const prewarmSteps = 40

// Model is the Bubble Tea model.
type Model struct {
	cfg       config.Config
	state     state
	report    *weather.Report
	err       error
	scene     scene.Scene
	params    scene.Params
	canvas    *scene.Canvas
	palette   scene.Palette
	styles    styles
	rng       *rand.Rand
	width     int // terminal size
	height    int
	layout    Layout
	showPanel bool
	ambient   bool
	showHelp  bool
	units     weather.Units
	autoUnits bool
	fetching  bool
	frame     int
	spinner   int

	demo *demoState // non-nil in --demo mode

	load   Loader
	ctx    context.Context
	cancel context.CancelFunc
	now    func() time.Time
}

// Options configure a Model.
type Options struct {
	Config  config.Config
	Palette scene.Palette // build before the program starts
	Loader  Loader        // unused in demo mode
	Rand    *rand.Rand    // nil = seeded from the clock
	Now     func() time.Time
}

// New builds the model.
func New(o Options) Model {
	rng := o.Rand
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}
	ctx, cancel := context.WithCancel(context.Background())
	units, ok := weather.ParseUnits(o.Config.Units)
	m := Model{
		cfg:       o.Config,
		state:     stateLoading,
		canvas:    scene.NewCanvas(0, 0),
		palette:   o.Palette,
		styles:    newStyles(o.Config.NoColor, o.Config.ASCII),
		rng:       rng,
		showPanel: true,
		ambient:   o.Config.Ambient,
		showHelp:  true,
		units:     units,
		autoUnits: !ok,
		load:      o.Loader,
		ctx:       ctx,
		cancel:    cancel,
		now:       now,
	}
	if o.Config.Demo {
		m.demo = newDemoState(now(), units)
		m.applyReport(m.demo.report())
	}
	return m
}

func (m Model) frameDuration() time.Duration {
	return time.Second / time.Duration(max(m.cfg.FPS, 1))
}

// Init starts the first fetch, the animation and the refresh timer.
func (m Model) Init() tea.Cmd {
	if m.demo != nil {
		return tick(m.frameDuration())
	}
	return tea.Batch(m.fetch(false), tick(m.frameDuration()), refreshTimer())
}

func (m *Model) fetch(bypassCache bool) tea.Cmd {
	m.fetching = true
	return fetchCmd(m.ctx, m.load, LoadRequest{Units: m.units, AutoUnits: m.autoUnits, BypassCache: bypassCache})
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg.String())

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.relayout()

	case tickMsg:
		m.frame++
		m.spinner++
		if m.scene != nil {
			m.scene.Step()
		}
		if m.demo != nil && m.demo.shouldAdvance(m.now()) {
			m.demo.move(1, m.now(), false)
			m.applyReport(m.demo.report())
		}
		return m, tick(m.frameDuration())

	case refreshMsg:
		return m, tea.Batch(m.fetch(false), refreshTimer())

	case weatherMsg:
		m.fetching = false
		m.err = nil
		m.units, m.autoUnits = msg.report.Units, false
		m.applyReport(msg.report)

	case errMsg:
		m.fetching = false
		m.err = msg.err
		if m.report != nil {
			// Keep showing the last report, flagged as offline.
			r := *m.report
			r.Stale = true
			m.report = &r
			m.units = r.Units // undo a failed units toggle
		} else {
			m.state = stateError
		}
	}
	return m, nil
}

func (m Model) handleKey(k string) (tea.Model, tea.Cmd) {
	switch {
	case matches(k, keysQuit):
		m.cancel()
		return m, tea.Quit
	case matches(k, keysHelp):
		m.showHelp = !m.showHelp
	case matches(k, keysPanel):
		m.showPanel = !m.showPanel
		m.relayout()
	case matches(k, keysAmbient):
		m.ambient = !m.ambient
		m.relayout()
	case m.demo != nil:
		return m.handleDemoKey(k)
	case matches(k, keysRefresh) && !m.fetching:
		if m.state == stateError {
			m.state = stateLoading
		}
		return m, m.fetch(true)
	case matches(k, keysUnits) && m.state == stateReady && !m.fetching:
		m.units, m.autoUnits = toggleUnits(m.units), false
		return m, m.fetch(false)
	}
	return m, nil
}

// handleDemoKey handles scene navigation and units in --demo mode; the
// demo never touches the network, so refresh does nothing.
func (m Model) handleDemoKey(k string) (tea.Model, tea.Cmd) {
	switch {
	case matches(k, keysNext):
		m.demo.move(1, m.now(), true)
	case matches(k, keysPrev):
		m.demo.move(-1, m.now(), true)
	case matches(k, keysUnits):
		m.demo.units = toggleUnits(m.demo.units)
	default:
		return m, nil
	}
	m.applyReport(m.demo.report())
	return m, nil
}

// sceneParams derives scene parameters from a report (or a forced --scene).
func (m *Model) sceneParams(r *weather.Report) scene.Params {
	c := r.Current
	cond, in, _ := weather.Classify(c.Code, c.IsDay)
	p := scene.Params{
		Condition:  cond,
		Intensity:  in,
		Wind:       weather.WindDrift(c.WindSpeed, c.WindDirDeg, r.Units),
		IsDay:      c.IsDay,
		ASCIIOnly:  m.cfg.ASCII,
		SnowGrains: weather.IsSnowGrains(c.Code),
	}
	if fc, fi, day, ok := scene.ByName(m.cfg.Scene); ok {
		p.Condition, p.Intensity, p.IsDay, p.SnowGrains = fc, fi, day, false
		if ci, ok := scene.ParseIntensity(m.cfg.Intensity); ok {
			p.Intensity = ci
		}
	}
	return p
}

// applyReport stores a report and rebuilds the scene if the weather changed.
func (m *Model) applyReport(r *weather.Report) {
	m.report = r
	m.state = stateReady
	p := m.sceneParams(r)
	if m.scene != nil && p == m.params {
		return // same weather: keep the running animation (and snow pile)
	}
	m.params = p
	m.scene = scene.New(p, m.rng)
	m.resizeScene()
}

func (m *Model) relayout() {
	l := ComputeLayout(m.width, m.height, m.showPanel, m.ambient)
	changed := l.SceneW != m.layout.SceneW || l.SceneH != m.layout.SceneH
	m.layout = l
	if changed {
		m.resizeScene()
	}
}

func (m *Model) resizeScene() {
	w, h := m.layout.SceneW, m.layout.SceneH
	m.canvas.Resize(w, h)
	if m.scene == nil || w == 0 || h == 0 {
		return
	}
	m.scene.Resize(w, h)
	for i := 0; i < prewarmSteps; i++ {
		m.scene.Step()
	}
}

func toggleUnits(u weather.Units) weather.Units {
	if u == weather.UnitsMetric {
		return weather.UnitsImperial
	}
	return weather.UnitsMetric
}
