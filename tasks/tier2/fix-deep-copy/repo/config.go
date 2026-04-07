package config

// Config represents application configuration.
type Config struct {
	Name     string
	Port     int
	Tags     []string
	Settings map[string]string
	Database *DatabaseConfig
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	Options  map[string]string
	Replicas []string
}

// NewConfig creates a new Config with defaults.
func NewConfig() *Config {
	return &Config{
		Name:     "default",
		Port:     8080,
		Tags:     []string{"prod"},
		Settings: map[string]string{"log_level": "info"},
		Database: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Options:  map[string]string{"sslmode": "disable"},
			Replicas: []string{"replica-1"},
		},
	}
}

// Clone creates a copy of the Config.
// BUG: This is a shallow copy - slices and maps share references.
func (c *Config) Clone() *Config {
	clone := *c
	// BUG: Tags slice is shared
	// BUG: Settings map is shared
	// BUG: Database pointer is shared
	return &clone
}
