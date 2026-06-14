package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestActiveHandlerServesRegisteredRouteWhenAllowedHostsEmpty(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterRoute("/custom", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("custom route reached"))
	})

	request := httptest.NewRequest(http.MethodGet, "/custom", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.Code)
	}
	if body := response.Body.String(); body != "custom route reached" {
		t.Fatalf("expected registered route body, got %q", body)
	}
}

func TestActiveHandlerServesRegisteredRouteWhenHostAllowed(t *testing.T) {
	server := NewGoServer(ServerConfig{
		AllowedHosts: []string{"example.com", "::1"},
	}, testLogger())
	server.RegisterRoute("/custom", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("allowed host reached"))
	})

	request := httptest.NewRequest(http.MethodGet, "/custom", nil)
	request.Host = "EXAMPLE.com:8081"
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, response.Code)
	}
	if body := response.Body.String(); body != "allowed host reached" {
		t.Fatalf("expected allowed route body, got %q", body)
	}
}

func TestActiveHandlerRejectsDisallowedHost(t *testing.T) {
	server := NewGoServer(ServerConfig{
		AllowedHosts: []string{"localhost"},
	}, testLogger())
	server.RegisterRoute("/custom", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("should not be reached"))
	})

	request := httptest.NewRequest(http.MethodGet, "/custom", nil)
	request.Host = "evil.example:8081"
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "Host Not Allowed") {
		t.Fatalf("expected host policy error page, got %q", body)
	}
}

func TestActiveHandlerServesEmbeddedCSS(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.AddDefaultGoServerRoutes()

	request := httptest.NewRequest(http.MethodGet, "/__go_server/static/styles.css", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "Base Styles") {
		t.Fatalf("expected embedded stylesheet body, got %q", body)
	}
}

func TestRegisterLocalSiteServesIndexHTML(t *testing.T) {
	siteDir := createTestSite(t)
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterLocalSite("/", siteDir, "index.html")
	server.AddDefaultGoServerRoutes()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "local site index") {
		t.Fatalf("expected index HTML body, got %q", body)
	}
}

func TestRegisterLocalSiteServesCSS(t *testing.T) {
	siteDir := createTestSite(t)
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterLocalSite("/", siteDir, "index.html")
	server.AddDefaultGoServerRoutes()

	request := httptest.NewRequest(http.MethodGet, "/styles.css", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "color: red") {
		t.Fatalf("expected stylesheet body, got %q", body)
	}
}

func TestRegisterLocalSiteMissingFileReturnsNotFound(t *testing.T) {
	siteDir := createTestSite(t)
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterLocalSite("/", siteDir, "index.html")
	server.AddDefaultGoServerRoutes()

	request := httptest.NewRequest(http.MethodGet, "/missing.css", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func createTestSite(t *testing.T) string {
	t.Helper()

	siteDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(siteDir, "index.html"), []byte("<h1>local site index</h1>"), 0644); err != nil {
		t.Fatalf("write index file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "styles.css"), []byte("body { color: red; }"), 0644); err != nil {
		t.Fatalf("write stylesheet: %v", err)
	}

	return siteDir
}
