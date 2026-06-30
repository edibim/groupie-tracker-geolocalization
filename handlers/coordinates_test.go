package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"groupie-tracker-geolocalization/geocode"
	"groupie-tracker-geolocalization/models"
)

func TestCoordinatesReturnsSortedJSON(t *testing.T) {
	// Populate the API relation cache for artist 1.
	cache.Relations.Index = []models.Relation{
		{ID: 1, DatesLocations: map[string][]string{
			"london-uk":       {"10-02-2020"},
			"los_angeles-usa": {"22-08-2019"},
		}},
	}

	// Inject a fake geocoder so the test never hits the network.
	geocoder = func(query string) (geocode.Coords, error) {
		switch query {
		case "London, UK":
			return geocode.Coords{Lat: 51.5, Lng: -0.12}, nil
		case "Los Angeles, USA":
			return geocode.Coords{Lat: 34.0, Lng: -118.2}, nil
		}
		return geocode.Coords{}, fmt.Errorf("unexpected query %q", query)
	}

	req := httptest.NewRequest(http.MethodGet, "/artist/1/coordinates", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	coordinatesHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got []struct {
		Address string   `json:"address"`
		Lat     float64  `json:"lat"`
		Lng     float64  `json:"lng"`
		Dates   []string `json:"dates"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d items, want 2", len(got))
	}
	// Earliest date first: LA (22-08-2019) before London (10-02-2020).
	if got[0].Address != "Los Angeles, USA" {
		t.Errorf("first = %q, want %q", got[0].Address, "Los Angeles, USA")
	}
	if got[1].Address != "London, UK" {
		t.Errorf("second = %q, want %q", got[1].Address, "London, UK")
	}
	if got[0].Lat != 34.0 {
		t.Errorf("first lat = %v, want 34.0", got[0].Lat)
	}
}

func TestCoordinatesRejectsNonGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/artist/1/coordinates", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	coordinatesHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestCoordinatesSkipsAndLogsFailedGeocoding(t *testing.T) {
	cache.Relations.Index = []models.Relation{
		{ID: 1, DatesLocations: map[string][]string{
			"london-uk":       {"10-02-2020"},
			"los_angeles-usa": {"22-08-2019"},
		}},
	}

	// London fails to geocode; Los Angeles succeeds.
	geocoder = func(query string) (geocode.Coords, error) {
		if query == "London, UK" {
			return geocode.Coords{}, fmt.Errorf("nominatim unreachable")
		}
		return geocode.Coords{Lat: 34.0, Lng: -118.2}, nil
	}

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	req := httptest.NewRequest(http.MethodGet, "/artist/1/coordinates", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	coordinatesHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var got []struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Address != "Los Angeles, USA" {
		t.Fatalf("got %+v, want only Los Angeles, USA", got)
	}
	if !strings.Contains(logBuf.String(), "London, UK") {
		t.Errorf("expected a warning logging the failed location, got %q", logBuf.String())
	}
}

func TestCoordinatesUnknownArtistReturns404(t *testing.T) {
	cache.Relations.Index = []models.Relation{} // no artists

	req := httptest.NewRequest(http.MethodGet, "/artist/999/coordinates", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	coordinatesHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
