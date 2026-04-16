package docgen

import "time"

type Config struct {
	Timeout time.Duration
}

type Option func(*Config)

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = timeout
	}
}

func defaultConfig() Config {
	return Config{Timeout: 10 * time.Second}
}
