package config

import "os"

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic("missing required environment variable: " + name)
	}
	return v
}

// Config contains app config
type Config struct {
	Loglevel       string
	BackoffMaxTime string
	DockerHost     string
}

// NewConfig returns a new app config
func NewConfig() *Config {
	return &Config{
		Loglevel:       getenv("LOGLEVEL"),
		BackoffMaxTime: getenv("BACKOFF_MAX_TIME"),
		DockerHost:     getenv("DOCKER_HOST"),
	}
}
