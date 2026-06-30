package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetArtistsReturnsData(t *testing.T) {
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":1,"name":"Queen","members":["Freddie Mercury"],"creationDate":1970,"firstAlbum":"14-12-1973"}]`))
	}))
	defer fake.Close()

	originalBase := baseURL
	baseURL = fake.URL
	defer func() { baseURL = originalBase }()

	artists, err := GetArtists()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(artists) == 0 {
		t.Fatal("expected artists, got empty slice")
	}
	if artists[0].Name != "Queen" {
		t.Fatalf("expected Queen, got %s", artists[0].Name)
	}
}

func TestGetArtistsFailsOnBadStatus(t *testing.T) {
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fake.Close()

	originalBase := baseURL
	baseURL = fake.URL
	defer func() { baseURL = originalBase }()

	_, err := GetArtists()
	if err == nil {
		t.Fatal("expected error on 500 status, got nil")
	}
}
