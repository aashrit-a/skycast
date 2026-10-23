package weather

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func serve(t *testing.T, status int, body []byte, hits *int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			*hits++
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "skycast/") {
			t.Errorf("missing User-Agent, got %q", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func testClient(base string, cache *Cache) *Client {
	c := NewClient("skycast/test", cache)
	c.BaseURL = base
	return c
}

func TestFetchRain(t *testing.T) {
	srv := serve(t, 200, fixture(t, "forecast_rain.json"), nil)
	r, err := testClient(srv.URL, nil).Fetch(context.Background(), 35.69, 139.69, UnitsMetric)
	if err != nil {
		t.Fatal(err)
	}
	c := r.Current
	if c.Temp != 18.2 || c.FeelsLike != 16.4 || c.Humidity != 82 || c.Code != 61 || !c.IsDay || c.WindDirDeg != 45 || c.WindSpeed != 12.3 {
		t.Errorf("current = %+v", c)
	}
	if len(r.Daily) != 4 {
		t.Fatalf("daily len = %d", len(r.Daily))
	}
	d0 := r.Daily[0]
	if d0.Sunrise.Format("15:04") != "05:58" || d0.Sunset.Format("15:04") != "17:02" {
		t.Errorf("sunrise/sunset = %v/%v", d0.Sunrise, d0.Sunset)
	}
	if _, off := d0.Sunrise.Zone(); off != 9*3600 {
		t.Errorf("offset = %d", off)
	}
	if r.Daily[1].Code != 3 || r.Daily[1].High != 21.3 || r.Daily[1].PrecipChance != 40 || r.Daily[1].Date.Weekday() != time.Friday {
		t.Errorf("daily[1] = %+v", r.Daily[1])
	}
	if r.Stale {
		t.Error("should not be stale")
	}
}

func TestFetchShortDailyWithNulls(t *testing.T) {
	srv := serve(t, 200, fixture(t, "forecast_clear_night.json"), nil)
	r, err := testClient(srv.URL, nil).Fetch(context.Background(), 40.71, -73.99, UnitsImperial)
	if err != nil {
		t.Fatal(err)
	}
	if r.Current.IsDay || r.Current.Code != 0 {
		t.Errorf("current = %+v", r.Current)
	}
	if len(r.Daily) != 3 {
		t.Fatalf("daily len = %d", len(r.Daily))
	}
	if r.Daily[2].Code != 3 || r.Daily[1].PrecipChance != 0 {
		t.Errorf("null handling: %+v", r.Daily)
	}
}

func TestFetchNon200(t *testing.T) {
	srv := serve(t, 503, []byte("oops"), nil)
	_, err := testClient(srv.URL, nil).Fetch(context.Background(), 1, 2, UnitsMetric)
	if err == nil || !strings.Contains(err.Error(), "503") {
		t.Fatalf("want 503 error, got %v", err)
	}
}

func TestFetchBadJSON(t *testing.T) {
	for _, body := range []string{"not json", `{"daily":{}}`} {
		srv := serve(t, 200, []byte(body), nil)
		if _, err := testClient(srv.URL, nil).Fetch(context.Background(), 1, 2, UnitsMetric); err == nil {
			t.Errorf("%q: want error", body)
		}
	}
}

func TestFetchUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	_, err := testClient(srv.URL, nil).Fetch(context.Background(), 1, 2, UnitsMetric)
	if !errors.Is(err, ErrUnreachable) {
		t.Fatalf("want ErrUnreachable, got %v", err)
	}
	if err.Error() != ErrUnreachable.Error() {
		t.Errorf("message = %q", err)
	}
}

func TestBuildForecastURL(t *testing.T) {
	for _, tt := range []struct {
		u                 Units
		temp, wind, precp string
	}{
		{UnitsMetric, "celsius", "kmh", "mm"},
		{UnitsImperial, "fahrenheit", "mph", "inch"},
	} {
		raw := buildForecastURL("", 48.85, 2.35, tt.u)
		p, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if p.Host != "api.open-meteo.com" || p.Path != "/v1/forecast" {
			t.Errorf("bad base: %s", raw)
		}
		q := p.Query()
		want := map[string]string{
			"latitude": "48.8500", "longitude": "2.3500", "timezone": "auto", "forecast_days": "4",
			"temperature_unit": tt.temp, "wind_speed_unit": tt.wind, "precipitation_unit": tt.precp,
			"current": "temperature_2m,apparent_temperature,relative_humidity_2m,is_day,weather_code,wind_speed_10m,wind_direction_10m,precipitation",
			"daily":   "weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,precipitation_probability_max",
		}
		for k, v := range want {
			if q.Get(k) != v {
				t.Errorf("%v: %s = %q, want %q", tt.u, k, q.Get(k), v)
			}
		}
	}
}
