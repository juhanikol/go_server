package main

import (
	"log/slog"
	"os"

	"go_server/pkg/httpserver"
)

const address = ":8083"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httpserver.NewGoServer(httpserver.ServerConfig{
		ServerAddress: address,
	}, logger)

	server.RegisterLocalSite("/", "examples/html_site/site", "index.html")
	server.AddDefaultGoServerRoutes()

	logger.Info("HTML site example listening", "addr", address, "url", "http://localhost:8083")
	if err := server.Start(); err != nil {
		logger.Error("HTML site example stopped", "error", err)
		os.Exit(1)
	}
}
