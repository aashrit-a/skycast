package location

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func server(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

func TestGeocodeSuccess(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "testdata", "geocode_tokyo.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gotName string
	srv := server(t, func(w http.ResponseWriter, r *http.Request) {
		gotName = r.URL.Query().Get("name")
		_, _ = w.Write(body)
	})
	res := &Resolver{GeocodeURL: srv.URL, UserAgent: "skycast/test"}
	loc, err := res.Geocode(context.Background(), "tokyo")
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "tokyo" {
		t.Errorf("name param = %q", gotName)
	}
	if loc.Display() != "Tokyo, Japan" || loc.CountryCode != "JP" || loc.Lat != 35.6895 || loc.Timezone != "Asia/Tokyo" {
		t.Errorf("loc = %+v", loc)
	}
}

func TestGeocodeEmpty(t *testing.T) {
	for _, body := range []string{`{"generationtime_ms":0.1}`, `{"results":[]}`} {
		srv := server(t, func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) })
		res := &Resolver{GeocodeURL: srv.URL}
		_, err := res.Geocode(context.Background(), "atlantys")
		if !errors.Is(err, ErrLocationNotFound) {
			t.Fatalf("want ErrLocationNotFound, got %v", err)
		}
		if err.Error() != `could not find a place called "atlantys"` {
			t.Errorf("message = %q", err)
		}
	}
}

func TestGeocodeQualifier(t *testing.T) {
	body := `{"results":[
		{"name":"San Jose","latitude":9.93,"longitude":-84.08,"country":"Costa Rica","country_code":"CR","admin1":"San José"},
		{"name":"San Jose","latitude":37.33,"longitude":-121.89,"country":"United States","country_code":"US","admin1":"California"}]}`
	var gotName string
	srv := server(t, func(w http.ResponseWriter, r *http.Request) {
		gotName = r.URL.Query().Get("name")
		_, _ = w.Write([]byte(body))
	})
	res := &Resolver{GeocodeURL: srv.URL}
	loc, err := res.Geocode(context.Background(), "san jose, ca")
	if err != nil {
		t.Fatal(err)
	}
	if gotName != "san jose" || loc.CountryCode != "US" {
		t.Errorf("name=%q loc=%+v", gotName, loc)
	}
	if _, err := res.Geocode(context.Background(), "san jose, zz"); !errors.Is(err, ErrLocationNotFound) {
		t.Errorf("want not found, got %v", err)
	}
}

func TestIPLookup(t *testing.T) {
	srv := server(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","country":"Japan","countryCode":"JP","regionName":"Tokyo","city":"Shinjuku","lat":35.69,"lon":139.7,"timezone":"Asia/Tokyo"}`))
	})
	loc, err := (&Resolver{IPURL: srv.URL}).FromIP(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if loc.Display() != "Shinjuku, Japan" || loc.CountryCode != "JP" {
		t.Errorf("loc = %+v", loc)
	}
}

func TestIPLookupFailure(t *testing.T) {
	fail := server(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"fail","message":"private range"}`))
	})
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()
	for _, u := range []string{fail.URL, down.URL} {
		_, err := (&Resolver{IPURL: u}).FromIP(context.Background())
		if !errors.Is(err, ErrIPLookup) {
			t.Fatalf("want ErrIPLookup, got %v", err)
		}
		if err.Error() != `couldn't detect your location; try: skycast "new york"` {
			t.Errorf("message = %q", err)
		}
	}
}

func TestResolveOrder(t *testing.T) {
	res := &Resolver{GeocodeURL: "http://invalid.invalid", IPURL: "http://invalid.invalid"}
	loc, err := res.Resolve(context.Background(), "ignored", 48.85, 2.35, true)
	if err != nil {
		t.Fatal(err)
	}
	if loc.Lat != 48.85 || loc.Lon != 2.35 || loc.Display() != "48.85, 2.35" {
		t.Errorf("loc = %+v", loc)
	}
}

func TestGeocodeErrors(t *testing.T) {
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()
	if _, err := (&Resolver{GeocodeURL: down.URL}).Geocode(context.Background(), "x"); !errors.Is(err, ErrGeocodeUnreachable) {
		t.Errorf("want ErrGeocodeUnreachable, got %v", err)
	}
	bad := server(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(502) })
	if _, err := (&Resolver{GeocodeURL: bad.URL}).Geocode(context.Background(), "x"); err == nil || !strings.Contains(err.Error(), "502") {
		t.Errorf("want HTTP 502 error, got %v", err)
	}
}
