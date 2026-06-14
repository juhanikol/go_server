package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"go_server/pkg/httpserver"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httpserver.NewGoServer(httpserver.ServerConfig{
		ServerAddress: ":8082",
	}, logger)

	server.RegisterRoute("/hello", http.MethodGet, func(responseWriter http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintln(responseWriter, "hello from minimal example")
	})
	server.AddDefaultGoServerRoutes()

	logger.Info("minimal example listening", "addr", ":8082", "route", "/hello")
	if err := server.Start(); err != nil {
		logger.Error("minimal example stopped", "error", err)
		os.Exit(1)
	}
}
