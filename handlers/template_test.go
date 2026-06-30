package handlers

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"groupie-tracker-geolocalization/models"
)

// TestHomeTemplateRendersArtists checks that the homepage renders with artist data.
func TestHomeTemplateRendersArtists(t *testing.T) {
	cache.Artists = []models.Artist{
		{ID: 1, Name: "Queen", Image: "queen.jpg", CreationDate: 1970},
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Queen") {
		t.Fatal("expected 'Queen' in response body")
	}
}

// TestRender404ReturnsNotFound checks that render404 returns status 404 and the error page.
func TestRender404ReturnsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()

	render404(rec)

	if rec.Code != 404 {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "404") {
		t.Fatal("expected '404' in response body")
	}
}

// TestRender500ReturnsServerError checks that render500 returns status 500 and the error page.
func TestRender500ReturnsServerError(t *testing.T) {
	rec := httptest.NewRecorder()

	render500(rec)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "500") {
		t.Fatal("expected '500' in response body")
	}
}

// TestRenderTemplateReturns500OnExecuteError checks failures that happen after
// a template parses successfully but cannot render the provided data.
func TestRenderTemplateReturns500OnExecuteError(t *testing.T) {
	tmplFile := filepath.Join(t.TempDir(), "broken.html")
	if err := os.WriteFile(tmplFile, []byte(`before {{.MissingField}}`), 0o600); err != nil {
		t.Fatalf("failed to create test template: %v", err)
	}

	rec := httptest.NewRecorder()
	renderTemplate(rec, tmplFile, struct{}{})

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "before") {
		t.Fatal("expected no partial template output before the 500 page")
	}
}

func TestRenderTemplateReturns500OnMissingFile(t *testing.T) {
	rec := httptest.NewRecorder()
	renderTemplate(rec, filepath.Join(t.TempDir(), "missing.html"), nil)

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "500") {
		t.Fatal("expected the 500 error page")
	}
}

// TestArtistTemplateRendersArtist checks that the artist page renders with artist data.
func TestArtistTemplateRendersArtist(t *testing.T) {
	cache.Artists = []models.Artist{
		{ID: 1, Name: "Queen", Image: "queen.jpg", CreationDate: 1970, FirstAlbum: "14-12-1973"},
	}

	req := httptest.NewRequest("GET", "/artist/1", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	artistHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Queen") {
		t.Fatal("expected 'Queen' in response body")
	}
}
