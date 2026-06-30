package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"groupie-tracker-geolocalization/api"
	"groupie-tracker-geolocalization/models"
)

var cache models.Cache

// cacheAPI fetches all 4 API endpoints concurrently using goroutines.
func cacheAPI() error {
	errCh := make(chan error, 4)

	go func() {
		var err error
		cache.Artists, err = api.GetArtists()
		errCh <- err
	}()

	go func() {
		var err error
		cache.Locations, err = api.GetLocations()
		errCh <- err
	}()

	go func() {
		var err error
		cache.Dates, err = api.GetDates()
		errCh <- err
	}()

	go func() {
		var err error
		cache.Relations, err = api.GetRelations()
		errCh <- err
	}()

	for i := 0; i < 4; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}

	return nil
}

// checkMethod reports whether the request uses GET.
// If not, it writes a 405 response and returns false so the caller can stop.
func checkMethod(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		render405(w)
		return false
	}
	return true
}

// RegisterRoutes connects project URLs to their handlers.
// main.go calls this once when the server is created.
func RegisterRoutes(mux *http.ServeMux) {
	if err := cacheAPI(); err != nil {
		fmt.Println("Error caching API", err)
		return
	}

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/artists", artistsHandler)
	mux.HandleFunc("/artist/{id}", artistHandler)
	mux.HandleFunc("/artist/{id}/members", artistMembersHandler)
	mux.HandleFunc("/artist/{id}/locations", artistLocationsHandler)
	mux.HandleFunc("/artist/{id}/dates", artistDatesHandler)
	mux.HandleFunc("/artist/{id}/relations", artistRelationsHandler)
	mux.HandleFunc("/artist/{id}/coordinates", coordinatesHandler)
	mux.HandleFunc("/locations", locationsHandler)
	mux.HandleFunc("/dates", datesHandler)
	mux.HandleFunc("/relation", relationHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// homeHandler serves the homepage template with artist data.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}

	if r.URL.Path != "/" {
		render404(w)
		return
	}
	renderTemplate(w, "templates/index.html", cache.Artists)
}

// getArtistByID finds an artist in the cache by ID and returns a pointer to it.
// Returns nil if no artist with that ID is found.
func getArtistByID(id int) *models.Artist {
	for i, artist := range cache.Artists {
		if artist.ID == id {
			return &cache.Artists[i]
		}
	}
	return nil
}

// artistsHandler returns all artists as JSON.
func artistsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	if r.URL.Path != "/artists" {
		render404(w)
		return
	}
	writeJSON(w, cache.Artists)
}

// artistHandler serves the artist page template for a single artist.
func artistHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}
	artist := getArtistByID(id)
	if artist == nil {
		render404(w)
		return
	}
	renderTemplate(w, "templates/artist.html", artist)
}

// artistMembersHandler returns the members of a single artist as JSON.
func artistMembersHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}
	artist := getArtistByID(id)
	if artist == nil {
		render404(w)
		return
	}
	writeJSON(w, artist.Members)
}

// artistLocationsHandler returns the locations of a single artist as JSON.
func artistLocationsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}
	for _, loc := range cache.Locations.Index {
		if loc.ID == id {
			writeJSON(w, loc.Locations)
			return
		}
	}
	render404(w)
}

// artistDatesHandler returns the concert dates of a single artist as JSON.
func artistDatesHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}
	for _, d := range cache.Dates.Index {
		if d.ID == id {
			writeJSON(w, d.Dates)
			return
		}
	}
	render404(w)
}

// artistRelationsHandler returns the date-location relations of a single artist as JSON.
func artistRelationsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		render404(w)
		return
	}
	for _, rel := range cache.Relations.Index {
		if rel.ID == id {
			writeJSON(w, rel.DatesLocations)
			return
		}
	}
	render404(w)
}

// locationsHandler returns all locations from cache as JSON.
func locationsHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	writeJSON(w, cache.Locations)
}

// datesHandler returns all dates from cache as JSON.
func datesHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	writeJSON(w, cache.Dates)
}

// relationHandler returns all relations from cache as JSON.
func relationHandler(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r) {
		return
	}
	writeJSON(w, cache.Relations)
}

// writeJSON encodes data as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
