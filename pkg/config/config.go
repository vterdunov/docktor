package config

import (
	"fmt"
	"os"
	"time"
)

// Config contains app config
type Config struct {
	JSONOutput     bool
	BackoffJitter  bool
	BackoffMinTime time.Duration
	BackoffMaxTime time.Duration
}

// NewConfig returns a new app config
func NewConfig() *Config {
	var jo bool
	_, exist := os.LookupEnv("JSON_OUTPUT")
	if exist {
		jo = true
	}

	var jitter bool
	_, exist = os.LookupEnv("BACKOFF_JITTER")
	if exist {
		jitter = true
	}

	mint := durationFromEnv("BACKOFF_MIN_TIME", 3*time.Second)
	maxt := durationFromEnv("BACKOFF_MAX_TIME", 30*time.Second)

	return &Config{
		JSONOutput:     jo,
		BackoffJitter:  jitter,
		BackoffMinTime: mint,
		BackoffMaxTime: maxt,
	}
}

func durationFromEnv(v string, t time.Duration) time.Duration {
	env, exist := os.LookupEnv(v)
	if exist {
		ut, err := time.ParseDuration(env)
		if err != nil {
			fmt.Printf("Could not convert BACKOFF_MIN_TIME to time. Will use default value: %d.\n", t)
			return time.Duration(t)
		}
		return time.Duration(ut)
	}
	return time.Duration(t)
}
