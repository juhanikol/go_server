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

	registerLocalSite(server, "examples/html_site/site", "index.html")
	server.AddDefaultGoServerRoutes()

	logger.Info("HTML site example listening", "addr", address, "url", "http://localhost:8083")
	if err := server.Start(); err != nil {
		logger.Error("HTML site example stopped", "error", err)
		os.Exit(1)
	}
}

func registerLocalSite(server *httpserver.GoServer, siteDir string, indexFile string) {
	if indexFile == "" {
		indexFile = "index.html"
	}

	fileServer := http.FileServer(http.Dir(siteDir))
	serveIndex := func(responseWriter http.ResponseWriter, request *http.Request) {
		http.ServeFile(responseWriter, request, filepath.Join(siteDir, indexFile))
	}

	server.SetHomeRoute(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			serveIndex(responseWriter, request)
			return
		}
		fileServer.ServeHTTP(responseWriter, request)
	})
}
