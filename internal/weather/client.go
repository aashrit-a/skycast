package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DefaultForecastURL is the Open-Meteo forecast endpoint.
const DefaultForecastURL = "https://api.open-meteo.com/v1/forecast"

// ErrUnreachable means the forecast service could not be contacted.
var ErrUnreachable = errors.New("can't reach the weather service — check your connection")

// NetError wraps a transport failure; it prints a friendly message and
// unwraps to both ErrUnreachable and the underlying error.
type NetError struct{ Err error }

func (e *NetError) Error() string { return ErrUnreachable.Error() }

// Unwrap returns ErrUnreachable and the underlying error.
func (e *NetError) Unwrap() []error { return []error{ErrUnreachable, e.Err} }

// Client fetches forecasts from Open-Meteo.
type Client struct {
	BaseURL     string       // defaults to DefaultForecastURL
	HTTP        *http.Client // defaults to a client with a 10s timeout
	UserAgent   string
	Cache       *Cache // nil disables caching
	NoCacheRead bool   // skip reading the cache (still writes)
	Now         func() time.Time
}

// NewClient returns a client with sensible defaults.
func NewClient(userAgent string, cache *Cache) *Client {
	return &Client{
		BaseURL:   DefaultForecastURL,
		HTTP:      &http.Client{Timeout: 10 * time.Second},
		UserAgent: userAgent,
		Cache:     cache,
	}
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// buildForecastURL returns the request URL for a location and unit system.
func buildForecastURL(base string, lat, lon float64, u Units) string {
	if base == "" {
		base = DefaultForecastURL
	}
	v := url.Values{}
	v.Set("latitude", strconv.FormatFloat(lat, 'f', 4, 64))
	v.Set("longitude", strconv.FormatFloat(lon, 'f', 4, 64))
	v.Set("current", "temperature_2m,apparent_temperature,relative_humidity_2m,is_day,weather_code,wind_speed_10m,wind_direction_10m,precipitation")
	v.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,precipitation_probability_max")
	v.Set("timezone", "auto")
	v.Set("forecast_days", "4")
	if u == UnitsImperial {
		v.Set("temperature_unit", "fahrenheit")
		v.Set("wind_speed_unit", "mph")
		v.Set("precipitation_unit", "inch")
	} else {
		v.Set("temperature_unit", "celsius")
		v.Set("wind_speed_unit", "kmh")
		v.Set("precipitation_unit", "mm")
	}
	return base + "?" + v.Encode()
}

// Fetch returns the report for a location. It serves a fresh cache entry
// without touching the network, and falls back to a stale entry (marked
// Stale) when the network fails. The Location field is left empty.
func (c *Client) Fetch(ctx context.Context, lat, lon float64, u Units) (*Report, error) {
	now := c.now()
	cached, savedAt, haveCache := c.Cache.Load(lat, lon, u)
	if c.NoCacheRead {
		haveCache = false
	}
	if haveCache && now.Sub(savedAt) < CacheTTL {
		if r, err := parseForecast(cached, u); err == nil {
			r.FetchedAt = savedAt
			return r, nil
		}
	}

	body, err := c.get(ctx, buildForecastURL(c.BaseURL, lat, lon, u))
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if haveCache && now.Sub(savedAt) < CacheMaxStale {
			if r, perr := parseForecast(cached, u); perr == nil {
				r.FetchedAt = savedAt
				r.Stale = true
				return r, nil
			}
		}
		return nil, err
	}
	r, err := parseForecast(body, u)
	if err != nil {
		return nil, err
	}
	c.Cache.Save(lat, lon, u, body, now)
	r.FetchedAt = now
	return r, nil
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, &NetError{Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, &NetError{Err: err}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather service returned HTTP %d", resp.StatusCode)
	}
	return body, nil
}

// apiResponse mirrors the parts of the Open-Meteo response we use. Numeric
// arrays use pointers because the API sends null for missing values.
type apiResponse struct {
	UTCOffsetSeconds int `json:"utc_offset_seconds"`
	Current          *struct {
		Temp      *float64 `json:"temperature_2m"`
		FeelsLike *float64 `json:"apparent_temperature"`
		Humidity  *float64 `json:"relative_humidity_2m"`
		IsDay     *int     `json:"is_day"`
		Code      *int     `json:"weather_code"`
		WindSpeed *float64 `json:"wind_speed_10m"`
		WindDir   *float64 `json:"wind_direction_10m"`
		Precip    *float64 `json:"precipitation"`
	} `json:"current"`
	Daily struct {
		Time         []string   `json:"time"`
		Code         []*int     `json:"weather_code"`
		High         []*float64 `json:"temperature_2m_max"`
		Low          []*float64 `json:"temperature_2m_min"`
		Sunrise      []string   `json:"sunrise"`
		Sunset       []string   `json:"sunset"`
		PrecipChance []*float64 `json:"precipitation_probability_max"`
	} `json:"daily"`
}

const apiTimeLayout = "2006-01-02T15:04"

func parseForecast(body []byte, u Units) (*Report, error) {
	var a apiResponse
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("unexpected response from weather service: %w", err)
	}
	if a.Current == nil || a.Current.Code == nil || a.Current.Temp == nil {
		return nil, errors.New("unexpected response from weather service: missing current conditions")
	}
	cur := a.Current
	r := &Report{
		Units: u,
		Current: Current{
			Temp:       *cur.Temp,
			FeelsLike:  derefF(cur.FeelsLike, *cur.Temp),
			Humidity:   int(derefF(cur.Humidity, 0)),
			WindSpeed:  derefF(cur.WindSpeed, 0),
			WindDirDeg: int(derefF(cur.WindDir, 0)),
			Precip:     derefF(cur.Precip, 0),
			IsDay:      cur.IsDay == nil || *cur.IsDay == 1,
			Code:       *cur.Code,
		},
	}

	// Times are local to the location without an offset; attach the
	// location's fixed offset so they display correctly.
	loc := time.FixedZone("", a.UTCOffsetSeconds)
	d := a.Daily
	for i := 0; i < len(d.Time) && i < 4; i++ {
		date, err := time.ParseInLocation("2006-01-02", d.Time[i], loc)
		if err != nil {
			return nil, fmt.Errorf("unexpected response from weather service: bad date %q", d.Time[i])
		}
		day := DailyForecast{
			Date:         date,
			Code:         derefI(at(d.Code, i), 3),
			High:         derefF(at(d.High, i), 0),
			Low:          derefF(at(d.Low, i), 0),
			PrecipChance: int(derefF(at(d.PrecipChance, i), 0)),
		}
		if i < len(d.Sunrise) {
			day.Sunrise, _ = time.ParseInLocation(apiTimeLayout, d.Sunrise[i], loc)
		}
		if i < len(d.Sunset) {
			day.Sunset, _ = time.ParseInLocation(apiTimeLayout, d.Sunset[i], loc)
		}
		r.Daily = append(r.Daily, day)
	}
	return r, nil
}

func at[T any](s []*T, i int) *T {
	if i < len(s) {
		return s[i]
	}
	return nil
}

func derefF(p *float64, def float64) float64 {
	if p == nil {
		return def
	}
	return *p
}

func derefI(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
