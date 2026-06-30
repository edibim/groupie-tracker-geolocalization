package models

// Artist represents a single artist/band from the API.
type Artist struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

// Location represents concert locations for one artist.
type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	Dates     string   `json:"dates"`
}

// Date represents concert dates for one artist.
type Date struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// Relation links concert dates to locations for one artist.
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// LocationsResponse is the top-level wrapper for /api/locations.
type LocationsResponse struct {
	Index []Location `json:"index"`
}

// DatesResponse is the top-level wrapper for /api/dates.
type DatesResponse struct {
	Index []Date `json:"index"`
}

// RelationsResponse is the top-level wrapper for /api/relation.
type RelationsResponse struct {
	Index []Relation `json:"index"`
}

// Cache holds all API data fetched once at server startup.
type Cache struct {
	Artists   []Artist
	Locations LocationsResponse
	Dates     DatesResponse
	Relations RelationsResponse
}
