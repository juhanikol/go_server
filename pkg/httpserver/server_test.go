package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
