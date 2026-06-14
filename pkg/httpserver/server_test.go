package httpserver

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewGoServerAppliesTimeouts(t *testing.T) {
	server := NewGoServer(ServerConfig{
		ServerAddress:           ":0",
		ServerReadTimeout:       10 * time.Second,
		ServerReadHeaderTimeout: 5 * time.Second,
		ServerWriteTimeout:      20 * time.Second,
		ServerIdleTimeout:       60 * time.Second,
	}, testLogger())

	if server.GoServerServing.ReadTimeout != 10*time.Second {
		t.Fatalf("expected read timeout, got %v", server.GoServerServing.ReadTimeout)
	}
	if server.GoServerServing.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read header timeout, got %v", server.GoServerServing.ReadHeaderTimeout)
	}
	if server.GoServerServing.WriteTimeout != 20*time.Second {
		t.Fatalf("expected write timeout, got %v", server.GoServerServing.WriteTimeout)
	}
	if server.GoServerServing.IdleTimeout != 60*time.Second {
		t.Fatalf("expected idle timeout, got %v", server.GoServerServing.IdleTimeout)
	}
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

func TestActiveHandlerRejectsWrongMethod(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterRoute("/post-only", http.MethodPost, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("should not be reached"))
	})

	request := httptest.NewRequest(http.MethodGet, "/post-only", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
	if allow := response.Header().Get("Allow"); allow != http.MethodPost {
		t.Fatalf("expected Allow header %q, got %q", http.MethodPost, allow)
	}
	if body := response.Body.String(); !strings.Contains(body, "Method Not Allowed") {
		t.Fatalf("expected method error body, got %q", body)
	}
}

func TestActiveHandlerRecoversFromPanic(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.RegisterRoute("/panic", http.MethodGet, func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	response := httptest.NewRecorder()

	server.activeHandler().ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "System Interruption") {
		t.Fatalf("expected recovery error page, got %q", body)
	}
}

func TestRenderGoServerErrorReturnsStatusAndPage(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	response := httptest.NewRecorder()

	server.RenderGoServerError(response, GoServerError{
		StatusCode: http.StatusTeapot,
		Title:      "Short And Stout",
		Message:    "The server refused coffee.",
	})

	if response.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Short And Stout") || !strings.Contains(body, "The server refused coffee.") {
		t.Fatalf("expected rendered error page, got %q", body)
	}
}

func TestDefaultRoutesServeRootStaticHealthAnd404(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.Manifest.StaticDir = filepath.Join("..", "..", "web", "static")
	server.SetHomeRoute(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("home page"))
	})
	server.AddDefaultGoServerRoutes()

	t.Run("root", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); body != "home page" {
			t.Fatalf("expected home page body, got %q", body)
		}
	})

	t.Run("health", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); body != "OK" {
			t.Fatalf("expected health body, got %q", body)
		}
	})

	t.Run("static", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/__go_server/static/styles.css", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, "Base Styles") {
			t.Fatalf("expected embedded static body, got %q", body)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, "Page Not Found") {
			t.Fatalf("expected not found error page, got %q", body)
		}
	})
}

func TestInternalHelpRoutesServeReadOnlyContent(t *testing.T) {
	server := NewGoServer(ServerConfig{}, testLogger())
	server.AddDefaultGoServerRoutes()

	t.Run("help", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/__go_server/help", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, "GoServer Help") {
			t.Fatalf("expected help page, got %q", body)
		}
	})

	t.Run("readme", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/__go_server/readme", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); !strings.Contains(body, "Server Help") {
			t.Fatalf("expected readme content, got %q", body)
		}
	})

	t.Run("internal-health", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/__go_server/health", nil)
		response := httptest.NewRecorder()

		server.activeHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
		}
		if body := response.Body.String(); body != "OK" {
			t.Fatalf("expected health body, got %q", body)
		}
	})
}
