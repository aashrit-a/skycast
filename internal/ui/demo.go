package ui

import (
	"time"

	"github.com/aashrit-a/skycast/internal/weather"
)

// Demo timing.
const (
	demoAdvance    = 6 * time.Second
	demoManualHold = 15 * time.Second
)

// demoScene is one synthetic report in the demo cycle (metric values).
type demoScene struct {
	code             int
	isDay            bool
	temp, feels      float64
	humidity         int
	wind             float64 // km/h
	windDir          int
	daily            [3]int // forecast codes
	highs, lows      [3]float64
	precip           [3]int
	todayHi, todayLo float64
}

// demoScenes covers every scene, in display order.
var demoScenes = []demoScene{
	{code: 0, isDay: true, temp: 24, feels: 25, humidity: 40, wind: 8, windDir: 200, daily: [3]int{1, 2, 0}, highs: [3]float64{26, 24, 27}, lows: [3]float64{15, 14, 16}, precip: [3]int{0, 10, 0}, todayHi: 25, todayLo: 14},
	{code: 0, isDay: false, temp: 12, feels: 11, humidity: 65, wind: 5, windDir: 300, daily: [3]int{0, 1, 2}, highs: [3]float64{20, 21, 19}, lows: [3]float64{10, 11, 9}, precip: [3]int{0, 0, 10}, todayHi: 19, todayLo: 10},
	{code: 2, isDay: true, temp: 19, feels: 19, humidity: 55, wind: 14, windDir: 250, daily: [3]int{3, 61, 2}, highs: [3]float64{18, 16, 20}, lows: [3]float64{11, 10, 12}, precip: [3]int{20, 60, 10}, todayHi: 20, todayLo: 11},
	{code: 3, isDay: true, temp: 15, feels: 14, humidity: 75, wind: 18, windDir: 230, daily: [3]int{61, 3, 2}, highs: [3]float64{14, 15, 17}, lows: [3]float64{9, 9, 10}, precip: [3]int{70, 30, 10}, todayHi: 16, todayLo: 10},
	{code: 45, isDay: true, temp: 8, feels: 6, humidity: 97, wind: 3, windDir: 90, daily: [3]int{45, 3, 2}, highs: [3]float64{11, 13, 14}, lows: [3]float64{5, 6, 7}, precip: [3]int{10, 20, 10}, todayHi: 10, todayLo: 5},
	{code: 53, isDay: true, temp: 13, feels: 12, humidity: 90, wind: 12, windDir: 220, daily: [3]int{61, 63, 3}, highs: [3]float64{14, 13, 15}, lows: [3]float64{9, 8, 9}, precip: [3]int{60, 80, 30}, todayHi: 14, todayLo: 9},
	{code: 65, isDay: true, temp: 14, feels: 10, humidity: 94, wind: 48, windDir: 260, daily: [3]int{63, 80, 2}, highs: [3]float64{15, 16, 18}, lows: [3]float64{10, 11, 11}, precip: [3]int{90, 70, 20}, todayHi: 15, todayLo: 10},
	{code: 67, isDay: true, temp: -1, feels: -6, humidity: 92, wind: 20, windDir: 20, daily: [3]int{73, 3, 71}, highs: [3]float64{0, 1, -1}, lows: [3]float64{-5, -4, -7}, precip: [3]int{80, 30, 50}, todayHi: 1, todayLo: -4},
	{code: 73, isDay: true, temp: -3, feels: -8, humidity: 88, wind: 15, windDir: 330, daily: [3]int{75, 71, 3}, highs: [3]float64{-2, -1, 0}, lows: [3]float64{-8, -7, -6}, precip: [3]int{90, 60, 20}, todayHi: -1, todayLo: -7},
	{code: 95, isDay: true, temp: 26, feels: 29, humidity: 80, wind: 35, windDir: 240, daily: [3]int{95, 80, 2}, highs: [3]float64{28, 27, 29}, lows: [3]float64{20, 19, 21}, precip: [3]int{80, 60, 20}, todayHi: 30, todayLo: 21},
	{code: 99, isDay: true, temp: 22, feels: 21, humidity: 85, wind: 42, windDir: 280, daily: [3]int{96, 95, 3}, highs: [3]float64{24, 25, 23}, lows: [3]float64{16, 17, 15}, precip: [3]int{85, 70, 30}, todayHi: 25, todayLo: 17},
}

// demoState tracks the --demo cycle.
type demoState struct {
	idx       int
	shownAt   time.Time // when the current scene appeared
	manualAt  time.Time // last ←/→ press
	units     weather.Units
	startDate time.Time
}

func newDemoState(now time.Time, units weather.Units) *demoState {
	return &demoState{shownAt: now, units: units, startDate: now}
}

// shouldAdvance reports whether it's time to auto-advance.
func (d *demoState) shouldAdvance(now time.Time) bool {
	if !d.manualAt.IsZero() && now.Sub(d.manualAt) < demoManualHold {
		return false
	}
	return now.Sub(d.shownAt) >= demoAdvance
}

// move steps the cycle by delta (wrapping); manual marks a key press.
func (d *demoState) move(delta int, now time.Time, manual bool) {
	n := len(demoScenes)
	d.idx = ((d.idx+delta)%n + n) % n
	d.shownAt = now
	if manual {
		d.manualAt = now
	}
}

// label is the footer text, e.g. "demo · 7/11 · Heavy rain".
func (d *demoState) label(st styles) string {
	s := demoScenes[d.idx]
	_, _, l := weather.Classify(s.code, s.isDay)
	return st.join("demo", itoa(d.idx+1)+"/"+itoa(len(demoScenes)), l)
}

func itoa(i int) string { return round(float64(i)) }

// report builds the synthetic report for the current scene.
func (d *demoState) report() *weather.Report {
	return demoReport(d.idx, d.units, d.startDate)
}

func demoReport(idx int, u weather.Units, now time.Time) *weather.Report {
	s := demoScenes[idx]
	temp := func(c float64) float64 {
		if u == weather.UnitsImperial {
			return c*9/5 + 32
		}
		return c
	}
	speed := s.wind
	if u == weather.UnitsImperial {
		speed = s.wind / 1.609344
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	r := &weather.Report{
		Location: "Demo City",
		Units:    u,
		Current: weather.Current{
			Temp: temp(s.temp), FeelsLike: temp(s.feels), Humidity: s.humidity,
			WindSpeed: speed, WindDirDeg: s.windDir, IsDay: s.isDay, Code: s.code,
		},
		FetchedAt: now,
	}
	r.Daily = append(r.Daily, weather.DailyForecast{
		Date: today, Code: s.code, High: temp(s.todayHi), Low: temp(s.todayLo),
		Sunrise: today.Add(6*time.Hour + 42*time.Minute), Sunset: today.Add(18*time.Hour + 9*time.Minute),
	})
	for i := 0; i < 3; i++ {
		day := today.AddDate(0, 0, i+1)
		r.Daily = append(r.Daily, weather.DailyForecast{
			Date: day, Code: s.daily[i], High: temp(s.highs[i]), Low: temp(s.lows[i]), PrecipChance: s.precip[i],
			Sunrise: day.Add(6*time.Hour + 43*time.Minute), Sunset: day.Add(18*time.Hour + 7*time.Minute),
		})
	}
	return r
}
