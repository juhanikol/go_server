package main

import (
	"go_server/cmd/example_server/myproject"
	"go_server/pkg/httpserver"
	"go_server/pkg/serverapp"
	"log/slog"
)

func main() {
	config := serverapp.AppConfig{
		LogFileName:   "app.log",
		LogLevel:      slog.LevelInfo,
		ServerAddress: ":8081",
		// The BuildServer callback is where we bridge the Manifest to the Server instance.
		BuildServer: func(Server *httpserver.GoServer) error {
			// 1. Retrieve the manifest from the project
			manifest := myproject.GetManifest()

			// 2. Populate the Server Manifest
			Server.Manifest = manifest

			// 3. Trigger the Phase 1 Scanner (Bulletproof Error check occurs here)
			Server.ScanProjectResources()

			// 4. Trigger the Interpreter to map the RouteMap
			Server.InterpretManifest()

			// CRITICAL: This registers the internal assets and the landing page fallback
			Server.AddDefaultGoServerRoutes()

			return nil
		},
	}

	// serverapp.Run handles the logging setup and signal listening automatically.
	if err := serverapp.Run(config); err != nil {
		panic(err)
	}
}
