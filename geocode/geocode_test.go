package geocode

import (
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"city and country", "mainz-germany", "Mainz, Germany"},
		{"multi-word city", "north_carolina-usa", "North Carolina, USA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.raw)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestGeocode(t *testing.T) {
	// Load a saved Nominatim response so the test never hits the network.
	body, err := os.ReadFile("testdata/nominatim_mainz.json")
	if err != nil {
		t.Fatal(err)
	}

	// Stand up a fake Nominatim server that returns the fixture.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Errorf("request has no User-Agent header")
		}
		w.Write(body)
	}))
	defer srv.Close()

	// Point the package at the fake server for the duration of the test.
	old := nominatimURL
	nominatimURL = srv.URL
	defer func() { nominatimURL = old }()

	lat, lng, err := Geocode("Mainz, Germany")
	if err != nil {
		t.Fatalf("Geocode returned error: %v", err)
	}
	if math.Abs(lat-49.9928617) > 0.0001 {
		t.Errorf("lat = %v, want ~49.9928617", lat)
	}
	if math.Abs(lng-8.2472526) > 0.0001 {
		t.Errorf("lng = %v, want ~8.2472526", lng)
	}
}

func TestGeocodeThrottlesRequests(t *testing.T) {
	body, err := os.ReadFile("testdata/nominatim_mainz.json")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer srv.Close()

	old := nominatimURL
	nominatimURL = srv.URL
	defer func() { nominatimURL = old }()

	// Small interval + reset so the test is fast and isolated.
	oldInterval := throttleInterval
	throttleInterval = 50 * time.Millisecond
	lastRequest = time.Time{}
	defer func() { throttleInterval = oldInterval }()

	start := time.Now()
	if _, _, err := Geocode("A"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Geocode("B"); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	if elapsed < throttleInterval {
		t.Errorf("two calls took %v, want >= %v (throttle not applied)", elapsed, throttleInterval)
	}
}
