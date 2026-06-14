package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"go_server/pkg/httpserver"
)

const address = ":8083"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httpserver.NewGoServer(httpserver.ServerConfig{
		ServerAddress: address,
	}, logger)

	siteDir := filepath.Join("examples", "html_site", "site")

	server.SetHomeRoute(func(responseWriter http.ResponseWriter, request *http.Request) {
		http.ServeFile(responseWriter, request, filepath.Join(siteDir, "index.html"))
	})
	server.RegisterRoute("/styles.css", http.MethodGet, func(responseWriter http.ResponseWriter, request *http.Request) {
		http.ServeFile(responseWriter, request, filepath.Join(siteDir, "styles.css"))
	})
	server.AddDefaultGoServerRoutes()

	logger.Info("HTML site example listening", "addr", address, "url", "http://localhost:8083")
	if err := server.Start(); err != nil {
		logger.Error("HTML site example stopped", "error", err)
		os.Exit(1)
	}
}
