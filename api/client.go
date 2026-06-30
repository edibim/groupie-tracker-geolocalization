// Package api handles all communication with the Groupie Tracker external API.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"groupie-tracker-geolocalization/models"
)

// baseURL is a var (not const) so tests can swap it with a fake server.
var baseURL = "https://groupietrackers.herokuapp.com/api"

func GetArtists() ([]models.Artist, error) {
	var artists []models.Artist
	err := fetch(baseURL+"/artists", &artists)
	return artists, err
}

func GetLocations() (models.LocationsResponse, error) {
	var locations models.LocationsResponse
	err := fetch(baseURL+"/locations", &locations)
	return locations, err
}

func GetDates() (models.DatesResponse, error) {
	var dates models.DatesResponse
	err := fetch(baseURL+"/dates", &dates)
	return dates, err
}

func GetRelations() (models.RelationsResponse, error) {
	var relations models.RelationsResponse
	err := fetch(baseURL+"/relation", &relations)
	return relations, err
}

// fetch makes a GET request to url and decodes the JSON response into target.
func fetch(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
