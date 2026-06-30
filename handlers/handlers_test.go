package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestMain runs handler tests from the project root.
// This keeps template and static file paths the same as go run .
func TestMain(m *testing.M) {
	originalDir, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}

	if err := os.Chdir(".."); err != nil {
		os.Exit(1)
	}

	code := m.Run()
	os.Chdir(originalDir)
	os.Exit(code)
}

// newTestServer creates a real test server with the project routes.
// It proves the routing setup can serve HTTP requests without starting port 8080.
func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	return httptest.NewServer(mux)
}

// TestServerStarts checks that a test server can start and answer.
// This is safer than binding the real localhost:8080 during tests.
func TestServerStarts(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("expected server to answer, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

// TestHomeRouteReturnsOK checks the homepage route.
// The homepage is the first route auditors usually open in the browser.
func TestHomeRouteReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

// TestArtistsRouteReturnsOK checks the artists placeholder route.
// Later milestones can replace the placeholder without changing the URL.
func TestArtistsRouteReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/artists", nil)
	rec := httptest.NewRecorder()

	artistsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

// TestInvalidRouteReturnsNotFound checks unknown routes.
// It protects the homepage handler from accepting every URL by mistake.
func TestInvalidRouteReturnsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

// TestArtistRejectsNonGet checks that a non-GET method is rejected on /artist/{id}.
// This route only reads data, so anything other than GET must return 405.
func TestArtistRejectsNonGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/artist/1", nil)
	rec := httptest.NewRecorder()

	artistHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

// TestHomeRejectsNonGetCleanly checks that a rejected method does not also
// serve the homepage body. The 405 response must not contain the artist grid.
func TestHomeRejectsNonGetCleanly(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if strings.Contains(string(body), "artist-grid") {
		t.Fatalf("405 response should not contain the homepage body")
	}
}
