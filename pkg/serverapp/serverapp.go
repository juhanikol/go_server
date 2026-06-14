// Package serverapp provides an optional convenience runner for GoServer.
//
// Purpose:
// This package is meant for users who want a fast and simple way to start
// a server project without rewriting the same boilerplate every time.
//
// It handles:
// - logger setup
// - server construction
// - route/module registration callback
// - default route installation
// - OS signal handling (Ctrl+C / SIGTERM)
// - graceful shutdown with timeout
//
// Advanced users can still ignore this package and use pkg/httpserver directly.
package serverapp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go_server/pkg/httpserver"
	"go_server/pkg/logging"
)

// AppConfig contains the main inputs needed to start a GoServer application.
//
// BuildServer is the key extension point:
// the user creates routes, homepage, modules, API handlers, etc. there.
type AppConfig struct {
	// Basic Server Settings
	ServerAddress string `json:"server_address"` // e.g., ":8081"
	Env           string `json:"env"`
	RootPath      string `json:"root_path"` // Absolute path to project root

	// Timeout Settings
	ReadTimeout     time.Duration `json:"read_timeout_sec"`     //
	WriteTimeout    time.Duration `json:"write_timeout_sec"`    //
	IdleTimeout     time.Duration `json:"idle_timeout_sec"`     // Time to keep idle connections open
	ShutdownTimeout time.Duration `json:"shutdown_timeout_sec"` //

	// Logging & Diagnostics
	LogFileName  string     `json:"log_file_name"`   //
	LogMaxSizeMB int64      `json:"log_max_size_mb"` // Max size in bytes before rotation
	LogLevel     slog.Level `json:"log_level"`       //

	// Resource Discovery (Scanner Inputs)
	TemplateDir string `json:"template_dir"` // Path to HTML templates
	StaticDir   string `json:"static_dir"`   // Path to project static assets

	// Security & Advanced - implemented later
	UseTLS       bool     // Enable HTTPS
	CertFile     string   // Path to SSL certificate
	KeyFile      string   // Path to SSL key
	AllowedHosts []string // List of valid hostnames for security

	// Callback
	BuildServer func(Server *httpserver.GoServer) error //
}

// Run starts the application using a simple managed lifecycle.
//
// Flow:
// 1. Apply defaults if needed
// 2. Setup logger
// 3. Create GoServer
// 4. Let the caller register their own routes/modules
// 5. Add GoServer default infrastructure routes
// 6. Start server in a goroutine
// 7. Wait for OS signal or startup/runtime error
// 8. Gracefully shut down with timeout
func Run(Config AppConfig) error {
	// 1. Attempt to load from 'server.json' and merge with defaults
	var err error
	Config, err = loadAndMergeConfig(Config)
	if err != nil {
		err = fmt.Errorf("failed loading config: %w", err)
		fmt.Printf("CRITICAL STARTUP ERROR: %v\n", err)
		return err
	}

	// 2. Critical Validation: Stop if core values are missing
	if err := validateCriticalConfig(Config); err != nil {
		fmt.Printf("CRITICAL STARTUP ERROR: %v\n", err)
		return err
	}

	// 3. Setup Logger with rotation
	LogSession, err := logging.SetupLogger(
		filepath.Join(Config.RootPath, Config.LogFileName),
		slog.Level(Config.LogLevel),
		Config.LogMaxSizeMB*1024*1024,
	)
	if err != nil {
		return err
	}
	defer LogSession.Writer.Close()

	AppLogger := logging.Get("SERVER-APP")
	AppLogger.Info("Application startup initiated", "env", Config.Env)

	// 4. Initialize Server with values from Config
	Server := httpserver.NewGoServer(httpserver.ServerConfig{
		ServerAddress:      Config.ServerAddress,
		ServerReadTimeout:  time.Duration(Config.ReadTimeout) * time.Second,
		ServerWriteTimeout: time.Duration(Config.WriteTimeout) * time.Second,
	}, AppLogger)

	// Set Environment and Manifest
	Server.Env = Config.Env
	Server.Manifest = httpserver.ProjectManifest{
		TemplateDir: filepath.Join(Config.RootPath, Config.TemplateDir),
		StaticDir:   filepath.Join(Config.RootPath, Config.StaticDir),
	}

	if Config.BuildServer != nil {
		if err := Config.BuildServer(Server); err != nil {
			AppLogger.Error("Server build failed", "error", err)
			return err
		}
	}

	// 5. Managed Lifecycle (Signals and Background Start)
	ExitSignal := make(chan os.Signal, 1)
	signal.Notify(ExitSignal, os.Interrupt, syscall.SIGTERM)
	ServerError := make(chan error, 1)

	go func() { ServerError <- Server.Start() }()

	select {
	case err := <-ServerError:
		if err != nil {
			AppLogger.Error("Server runtime error", "error", err)
			return err
		}
		return nil
	case sig := <-ExitSignal:
		AppLogger.Info("Shutdown signal captured", "signal", sig.String())
	}

	ShutdownCtx, Cancel := context.WithTimeout(context.Background(), time.Duration(Config.ShutdownTimeout)*time.Second)
	defer Cancel()
	return Server.Shutdown(ShutdownCtx)
}

func validateCriticalConfig(c AppConfig) error {
	if c.ServerAddress == "" {
		return fmt.Errorf("server_address is missing")
	}
	if c.TemplateDir == "" {
		return fmt.Errorf("template_dir is missing")
	}
	return nil
}

func loadAndMergeConfig(c AppConfig) (AppConfig, error) {
	// Try to read server.json from common locations to support zero-boilerplate structures.
	configPaths := []string{
		"server.json",        // Standard root placement
		"web/server.json",    // Encapsulated web folder placement
		"config/server.json", // Standard config folder placement
	}

	var data []byte
	configPath := ""
	for _, path := range configPaths {
		configData, err := os.ReadFile(path)
		if err == nil {
			data = configData
			configPath = path
			break
		}
		if !os.IsNotExist(err) {
			return c, fmt.Errorf("read %s: %w", path, err)
		}
	}

	if configPath != "" {
		var fileConfig AppConfig
		var loggingConfig struct {
			LogLevel *slog.Level `json:"log_level"`
		}

		if err := json.Unmarshal(data, &fileConfig); err != nil {
			return c, fmt.Errorf("parse %s: %w", configPath, err)
		}
		if err := json.Unmarshal(data, &loggingConfig); err != nil {
			return c, fmt.Errorf("parse %s log_level: %w", configPath, err)
		}

		// Apply file values if the struct field is not zero.
		if fileConfig.ServerAddress != "" {
			c.ServerAddress = fileConfig.ServerAddress
		}
		if fileConfig.Env != "" {
			c.Env = fileConfig.Env
		}
		if fileConfig.RootPath != "" {
			c.RootPath = fileConfig.RootPath
		}
		if fileConfig.TemplateDir != "" {
			c.TemplateDir = fileConfig.TemplateDir
		}
		if fileConfig.StaticDir != "" {
			c.StaticDir = fileConfig.StaticDir
		}
		if fileConfig.ReadTimeout != 0 {
			c.ReadTimeout = fileConfig.ReadTimeout
		}
		if fileConfig.WriteTimeout != 0 {
			c.WriteTimeout = fileConfig.WriteTimeout
		}
		if fileConfig.ShutdownTimeout != 0 {
			c.ShutdownTimeout = fileConfig.ShutdownTimeout
		}
		if fileConfig.LogFileName != "" {
			c.LogFileName = fileConfig.LogFileName
		}
		if fileConfig.LogMaxSizeMB != 0 {
			c.LogMaxSizeMB = fileConfig.LogMaxSizeMB
		}
		if loggingConfig.LogLevel != nil {
			c.LogLevel = *loggingConfig.LogLevel
		}
	}

	// Apply safe defaults for missing values.
	if c.Env == "" {
		c.Env = "development"
	}
	if c.LogFileName == "" {
		c.LogFileName = "goserver.log"
	}
	if c.LogMaxSizeMB == 0 {
		c.LogMaxSizeMB = 10
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 10
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 10
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 10
	}

	return c, nil
}
