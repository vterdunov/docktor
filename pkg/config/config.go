package config

import (
	"fmt"
	"os"
	"time"
)

// Config contains app config
type Config struct {
	JSONOutput     bool
	BackoffMinTime time.Duration
	BackoffMaxTime time.Duration
}

// NewConfig returns a new app config
func NewConfig() *Config {
	var jo bool
	_, ok := os.LookupEnv("JSON_OUTPUT")
	if ok {
		jo = true
	}

	var bmint time.Duration = 5
	mintime, ok := os.LookupEnv("BACKOFF_MIN_TIME")
	if !ok {
		t, err := time.ParseDuration(mintime)
		if err != nil {
			fmt.Printf("Could not convert BACKOFF_MIN_TIME to time. Will use default value: %d.", bmint)
		}
		bmint = t
	}

	var bmaxt time.Duration = 300
	maxtime, ok := os.LookupEnv("BACKOFF_MAX_TIME")
	if !ok {
		t, err := time.ParseDuration(maxtime)
		if err != nil {
			fmt.Printf("Could not convert BACKOFF_MAX_TIME to time. Will use default value: %d", bmaxt)
		}
		bmaxt = t
	}

	return &Config{
		JSONOutput:     jo,
		BackoffMinTime: bmint,
		BackoffMaxTime: bmaxt,
	}
}
