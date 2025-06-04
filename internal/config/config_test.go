package config

import (
	"os"
	"testing"

	"github.com/ihatemodels/alcatraz-rest/internal/observability"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          8080,
				},
				log: LogConfig{
					Level: "info",
					Type:  "json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          8080,
				},
				log: LogConfig{
					Level: "invalid",
					Type:  "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid log type",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          8080,
				},
				log: LogConfig{
					Level: "info",
					Type:  "invalid",
				},
			},
			wantErr: true,
			errMsg:  "invalid log type",
		},
		{
			name: "invalid port - too low",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          0,
				},
				log: LogConfig{
					Level: "info",
					Type:  "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          70000,
				},
				log: LogConfig{
					Level: "info",
					Type:  "json",
				},
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "empty listen address",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "",
					Port:          8080,
				},
				log: LogConfig{
					Level: "info",
					Type:  "json",
				},
			},
			wantErr: true,
			errMsg:  "listen address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestConfig_GetServerAddress(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "default config",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "0.0.0.0",
					Port:          8080,
				},
			},
			want: "0.0.0.0:8080",
		},
		{
			name: "localhost config",
			config: Config{
				Server: ServerConfig{
					ListenAddress: "127.0.0.1",
					Port:          3000,
				},
			},
			want: "127.0.0.1:3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetServerAddress()
			if got != tt.want {
				t.Errorf("GetServerAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_InitLogger(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedFormat observability.OutputFormat
		expectedLevel  observability.LogLevel
	}{
		{
			name: "json info config",
			config: Config{
				log: LogConfig{
					Level: "info",
					Type:  "json",
				},
			},
			expectedFormat: observability.FormatJSON,
			expectedLevel:  observability.LevelInfo,
		},
		{
			name: "console debug config",
			config: Config{
				log: LogConfig{
					Level: "debug",
					Type:  "console",
				},
			},
			expectedFormat: observability.FormatConsole,
			expectedLevel:  observability.LevelDebug,
		},
		{
			name: "invalid type defaults to console",
			config: Config{
				log: LogConfig{
					Level: "warn",
					Type:  "invalid",
				},
			},
			expectedFormat: observability.OutputFormat("invalid"),
			expectedLevel:  observability.LevelWarn,
		},
		{
			name: "invalid level uses as-is",
			config: Config{
				log: LogConfig{
					Level: "invalid",
					Type:  "console",
				},
			},
			expectedFormat: observability.FormatConsole,
			expectedLevel:  observability.LogLevel("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.setObservabilityConfig()
			if tt.config.Observability.Format != tt.expectedFormat {
				t.Errorf("Format = %v, want %v", tt.config.Observability.Format, tt.expectedFormat)
			}
			if tt.config.Observability.Level != tt.expectedLevel {
				t.Errorf("Level = %v, want %v", tt.config.Observability.Level, tt.expectedLevel)
			}
		})
	}
}

func TestLoadFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
		expected Config
	}{
		{
			name: "invalid yaml",
			yaml: `
server:
  listen_address: "127.0.0.1"
  port: not_a_number
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "config_test_*.yaml")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write test YAML
			if _, err := tmpfile.Write([]byte(tt.yaml)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Start with default config
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

			// Test loading
			err = loadFromYAML(config, tmpfile.Name())

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadFromYAML() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("loadFromYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check results
			if config.Server.ListenAddress != tt.expected.Server.ListenAddress {
				t.Errorf("ListenAddress = %v, want %v", config.Server.ListenAddress, tt.expected.Server.ListenAddress)
			}
			if config.Server.Port != tt.expected.Server.Port {
				t.Errorf("Port = %v, want %v", config.Server.Port, tt.expected.Server.Port)
			}
			if config.log.Level != tt.expected.log.Level {
				t.Errorf("LogLevel = %v, want %v", config.log.Level, tt.expected.log.Level)
			}
			if config.log.Type != tt.expected.log.Type {
				t.Errorf("LogType = %v, want %v", config.log.Type, tt.expected.log.Type)
			}

			// Check that observability config was set correctly
			expectedObsFormat := observability.OutputFormat(tt.expected.log.Type)
			expectedObsLevel := observability.LogLevel(tt.expected.log.Level)
			if config.Observability.Format != expectedObsFormat {
				t.Errorf("Observability.Format = %v, want %v", config.Observability.Format, expectedObsFormat)
			}
			if config.Observability.Level != expectedObsLevel {
				t.Errorf("Observability.Level = %v, want %v", config.Observability.Level, expectedObsLevel)
			}
		})
	}
}

func TestLoadFromYAML_NonExistentFile(t *testing.T) {
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

	// Test with non-existent file - should not error and keep defaults
	err := loadFromYAML(config, "non_existent_file.yaml")
	if err != nil {
		t.Errorf("loadFromYAML() with non-existent file should not error, got %v", err)
	}

	// Config should remain unchanged
	if config.Server.ListenAddress != DefaultListenAddress {
		t.Errorf("ListenAddress = %v, want %v", config.Server.ListenAddress, DefaultListenAddress)
	}
	if config.Server.Port != DefaultPort {
		t.Errorf("Port = %v, want %v", config.Server.Port, DefaultPort)
	}
}

func TestApplyFlags(t *testing.T) {
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

	// Test flags
	listenAddress := "127.0.0.1"
	port := 9000
	logLevel := "debug"
	logType := "console"

	applyFlags(config, &listenAddress, &port, &logLevel, &logType)

	if config.Server.ListenAddress != listenAddress {
		t.Errorf("ListenAddress = %v, want %v", config.Server.ListenAddress, listenAddress)
	}
	if config.Server.Port != port {
		t.Errorf("Port = %v, want %v", config.Server.Port, port)
	}
	if config.log.Level != logLevel {
		t.Errorf("LogLevel = %v, want %v", config.log.Level, logLevel)
	}
	if config.log.Type != logType {
		t.Errorf("LogType = %v, want %v", config.log.Type, logType)
	}
}

func TestApplyFlags_EmptyValues(t *testing.T) {
	originalConfig := &Config{
		Server: ServerConfig{
			ListenAddress: DefaultListenAddress,
			Port:          DefaultPort,
		},
		log: LogConfig{
			Level: DefaultLogLevel,
			Type:  DefaultLogType,
		},
	}

	config := *originalConfig

	// Test with empty flags - should not change config
	emptyString := ""
	emptyInt := 0

	applyFlags(&config, &emptyString, &emptyInt, &emptyString, &emptyString)

	if config.Server.ListenAddress != originalConfig.Server.ListenAddress {
		t.Errorf("ListenAddress = %v, want %v", config.Server.ListenAddress, originalConfig.Server.ListenAddress)
	}
	if config.Server.Port != originalConfig.Server.Port {
		t.Errorf("Port = %v, want %v", config.Server.Port, originalConfig.Server.Port)
	}
	if config.log.Level != originalConfig.log.Level {
		t.Errorf("LogLevel = %v, want %v", config.log.Level, originalConfig.log.Level)
	}
	if config.log.Type != originalConfig.log.Type {
		t.Errorf("LogType = %v, want %v", config.log.Type, originalConfig.log.Type)
	}
}
