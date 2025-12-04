package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is exporter's configuration parameters defined by ENV or execution keys.
type Config struct {
	APIKey                   string
	TaskIDs                  []int
	EngMPNames               bool
	ApiUpdateDelay           time.Duration
	ApiDataTimeStep          time.Duration
	RefreshInterval          time.Duration
	MaxAllowedStalenessSteps int
	RequestDelay             time.Duration
	RequestRetries           int
	MaxRequestsPerSecond     int
	ListenAddress            string
	LocationsFilePath        string
	LogLevel                 string
}

// New create exporter config.
func New() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.APIKey, "api-key", envString("API_KEY", ""), "API key for Ping-Admin")
	taskIDsStr := flag.String("task-ids", envString("TASK_IDS", ""), "Comma-separated list of task IDs")
	flag.BoolVar(&cfg.EngMPNames, "eng-mp-names", envBool("ENG_MP_NAMES", true), "Translate monitoring points (MP) names to English")
	flag.DurationVar(&cfg.ApiUpdateDelay, "api-update-delay", envDuration("API_UPDATE_DELAY", 4*time.Minute), "Fixed Ping-Admin API delay for new data update")
	flag.DurationVar(&cfg.ApiDataTimeStep, "api-data-time-step", envDuration("API_DATA_TIME_STEP", 3*time.Minute), "Fixed Ping-Admin API time between data points")
	flag.DurationVar(&cfg.RefreshInterval, "refresh-interval", envDuration("REFRESH_INTERVAL", 3*time.Minute), "Exporter's refresh interval")
	flag.IntVar(&cfg.MaxAllowedStalenessSteps, "max-allowed-staleness-steps", envInt("MAX_ALLOWED_STALENESS_STEPS", 3), "Maximum allowed staleness steps")
	flag.DurationVar(&cfg.RequestDelay, "request-delay", envDuration("REQUEST_DElAY", 2*time.Second), "Minimum delay before API request (will be set to random between this and doubled values)")
	flag.IntVar(&cfg.RequestRetries, "request-retries", envInt("REQUEST_RETRIES", 3), "Maximum number of retries for API requests")
	flag.IntVar(&cfg.MaxRequestsPerSecond, "max-requests-per-second", envInt("MAX_REQUESTS_PER_SECOND", 2), "Maximum number of API requests allowed per second")
	flag.StringVar(&cfg.ListenAddress, "listen-address", envString("LISTEN_ADDRESS", ":8080"), "Address to listen on for HTTP requests")
	flag.StringVar(&cfg.LocationsFilePath, "locations-file", envString("LOCATIONS_FILE", "locations.json"), "Path to the locations.json translation file")
	flag.StringVar(&cfg.LogLevel, "log-level", envString("LOG_LEVEL", "info"), "Log level (e.g., debug, info, warn, error)")

	flag.Parse()

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required, please set --api-key or API_KEY environment variable")
	}

	if *taskIDsStr == "" {
		return nil, fmt.Errorf("task IDs are required, please set --task-ids or TASK_IDS environment variable")
	}

	var err error
	cfg.TaskIDs, err = parseTaskIDs(*taskIDsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid task IDs format: %w", err)
	}

	return cfg, nil
}

// parseTaskIDs get ID for each task from TaskIDs.
func parseTaskIDs(taskIDsStr string) ([]int, error) {
	if taskIDsStr == "" {
		return nil, nil
	}
	parts := strings.Split(taskIDsStr, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}
		id, err := strconv.Atoi(trimmedPart)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid integer", trimmedPart)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// envString string env variables helper.
func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// envDuration duration env variables helper.
func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

// envBool bool env variables helper.
func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// envInt int env variables helper.
func envInt(env string, def int) int {
	if v, ok := os.LookupEnv(env); ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return def
		}
		return i
	}
	return def
}
