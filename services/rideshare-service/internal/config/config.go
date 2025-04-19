package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig holds server-related settings
type ServerConfig struct {
	Port         string `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	LogLevel     string `yaml:"log_level"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Primary        DBConnection   `yaml:"primary"`
	Replicas       []DBConnection `yaml:"replicas"`
	MaxOpenConns   int            `yaml:"max_open_conns"`
	MaxIdleConns   int            `yaml:"max_idle_conns"`
	ConnMaxLifetime int           `yaml:"conn_max_lifetime"`
}

// DBConnection contains details for a database connection
type DBConnection struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// LoadConfig loads configuration from a YAML file and overrides with environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Initialize with defaults
	cfg := &Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  30,
			WriteTimeout: 30,
			LogLevel:     "info",
		},
		Database: DatabaseConfig{
			Primary: DBConnection{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "postgres",
				DBName:   "rideshare",
				SSLMode:  "disable",
			},
			MaxOpenConns:   25,
			MaxIdleConns:   5,
			ConnMaxLifetime: 300, // 5 minutes
		},
	}

	// Look for config file
	configFile := findConfigFile(configPath)
	
	// If found, load from file
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("error unmarshaling yaml: %w", err)
		}
	}

	// Override with environment variables
	applyEnvOverrides(cfg)

	return cfg, nil
}

// findConfigFile searches for a config file in the given directory and parent directories
func findConfigFile(startPath string) string {
	configFiles := []string{
		"config.yaml", 
		"config.yml",
	}

	// Check in the specified directory
	for _, filename := range configFiles {
		path := filepath.Join(startPath, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check in parent directory (up to 3 levels)
	parentPath := filepath.Dir(startPath)
	if parentPath != startPath && len(filepath.SplitList(parentPath)) > 1 {
		return findConfigFile(parentPath)
	}

	return ""
}

// applyEnvOverrides applies environment variable overrides to the config
func applyEnvOverrides(cfg *Config) {
	// Server settings
	if port := os.Getenv("SERVER_PORT"); port != "" {
		cfg.Server.Port = port
	}
	if timeout := getEnvInt("SERVER_READ_TIMEOUT", 0); timeout > 0 {
		cfg.Server.ReadTimeout = timeout
	}
	if timeout := getEnvInt("SERVER_WRITE_TIMEOUT", 0); timeout > 0 {
		cfg.Server.WriteTimeout = timeout
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Server.LogLevel = logLevel
	}

	// Database primary connection
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Primary.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		cfg.Database.Primary.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Primary.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Primary.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.Primary.DBName = dbName
	}
	if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
		cfg.Database.Primary.SSLMode = sslMode
	}

	// Database pool settings
	if maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 0); maxOpenConns > 0 {
		cfg.Database.MaxOpenConns = maxOpenConns
	}
	if maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 0); maxIdleConns > 0 {
		cfg.Database.MaxIdleConns = maxIdleConns
	}
	if connMaxLifetime := getEnvInt("DB_CONN_MAX_LIFETIME", 0); connMaxLifetime > 0 {
		cfg.Database.ConnMaxLifetime = connMaxLifetime
	}
}

// getEnvInt gets an environment variable as an integer
func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}