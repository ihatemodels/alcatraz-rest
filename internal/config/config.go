package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ihatemodels/alcatraz-live/internal/observability"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig `yaml:"server"`
	log    LogConfig    `yaml:"log"`

	// Observability configuration
	// set from the config values
	Observability observability.Config
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	ListenAddress string `yaml:"listen_address"`
	Port          int    `yaml:"port"`
}

// LogConfig holds logging-related configuration
type LogConfig struct {
	Level string `yaml:"level"`
	Type  string `yaml:"type"`
}

// Default configuration values
const (
	DefaultListenAddress = "0.0.0.0"
	DefaultPort          = 8080
	DefaultLogLevel      = "info"
	DefaultLogType       = "json"
)

// LoadConfig loads configuration from YAML file and command line flags
// Command line flags take precedence over YAML file values
func LoadConfig() (*Config, error) {
	// Define command line flags
	var (
		configFile    = flag.String("config", "config.yaml", "Path to configuration file")
		listenAddress = flag.String("listen-address", "", "Server listen address")
		port          = flag.Int("port", 0, "Server port")
		logLevel      = flag.String("log-level", "", "Log level (debug, info, warn, error)")
		logType       = flag.String("log-type", "", "Log type (console, json)")
		help          = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Start with default configuration
	config := &Config{
		Server: ServerConfig{
			ListenAddress: DefaultListenAddress,
			Port:          DefaultPort,
		},
		log: LogConfig{
			Level: DefaultLogLevel,
			Type:  DefaultLogType,
		},
	}

	// Load from YAML file if it exists
	if err := loadFromYAML(config, *configFile); err != nil {
		return nil, fmt.Errorf("failed to load config from YAML: %w", err)
	}

	// Override with command line flags if provided
	applyFlags(config, listenAddress, port, logLevel, logType)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromYAML loads configuration from a YAML file
func loadFromYAML(c *Config, filename string) error {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File doesn't exist, use defaults
		return nil
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	c.setObservabilityConfig()

	return nil
}

// applyFlags applies command line flags to the configuration
func applyFlags(c *Config, listenAddress *string, port *int, logLevel *string, logType *string) {
	if *listenAddress != "" {
		c.Server.ListenAddress = *listenAddress
	}
	if *port != 0 {
		c.Server.Port = *port
	}
	if *logLevel != "" {
		c.log.Level = *logLevel
	}
	if *logType != "" {
		c.log.Type = *logType
	}
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.log.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.log.Level)
	}

	// Validate log type
	validLogTypes := map[string]bool{
		"console": true,
		"json":    true,
	}
	if !validLogTypes[c.log.Type] {
		return fmt.Errorf("invalid log type: %s (must be console or json)", c.log.Type)
	}

	// Validate port
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	// Validate listen address (basic check)
	if c.Server.ListenAddress == "" {
		return fmt.Errorf("listen address cannot be empty")
	}

	return nil
}

// GetServerAddress returns the full server address (host:port)
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.ListenAddress, c.Server.Port)
}

func (c *Config) setObservabilityConfig() {
	c.Observability.Format = observability.OutputFormat(c.log.Type)
	c.Observability.Level = observability.LogLevel(c.log.Level)
	c.Observability.Writer = os.Stdout
}
