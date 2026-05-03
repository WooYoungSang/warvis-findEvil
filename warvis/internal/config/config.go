package config

import (
	"os"
	"strconv"
	"time"
)

// TimeoutConfig holds all timeout settings for Hunt operations.
type TimeoutConfig struct {
	ToolTimeout  time.Duration
	StateTimeout time.Duration
	HuntTimeout  time.Duration
}

// DefaultTimeoutConfig returns standard timeout values.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		ToolTimeout:  120 * time.Second,  // Max time per tool call
		StateTimeout: 600 * time.Second,  // Max time in any state
		HuntTimeout:  1800 * time.Second, // Max time for entire hunt (30 minutes)
	}
}

// LoadTimeoutConfig loads timeout values from environment variables.
// Falls back to defaults if env vars not set.
func LoadTimeoutConfig() TimeoutConfig {
	cfg := DefaultTimeoutConfig()

	if val := os.Getenv("WARVIS_TOOL_TIMEOUT_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			cfg.ToolTimeout = time.Duration(seconds) * time.Second
		}
	}

	if val := os.Getenv("WARVIS_STATE_TIMEOUT_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			cfg.StateTimeout = time.Duration(seconds) * time.Second
		}
	}

	if val := os.Getenv("WARVIS_HUNT_TIMEOUT_SECONDS"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			cfg.HuntTimeout = time.Duration(seconds) * time.Second
		}
	}

	return cfg
}
