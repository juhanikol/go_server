package serverapp

import (
	"encoding/json"
	"os"
)

// GSConfig matches the JSON structure for server settings.
type GSConfig struct {
	ServerAddress   string `json:"server_address"`
	RootPath        string `json:"root_path"` // Anchor for all relative paths
	LogFileName     string `json:"log_file_name"`
	LogMaxSize      int64  `json:"log_max_size"`
	LogLevel        int    `json:"log_level"`        // 0=Info, 4=Warn, 8=Error
	ShutdownTimeout int    `json:"shutdown_timeout"` // In seconds
}

// LoadConfig attempts to find and parse 'server.json'.
func LoadConfig() (GSConfig, error) {
	var config GSConfig

	// Default: Look in current working directory
	data, err := os.ReadFile("server.json")
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, err
	}

	// Logic: If RootPath is empty, use the current directory
	if config.RootPath == "" {
		config.RootPath, _ = os.Getwd()
	}

	return config, nil
}
