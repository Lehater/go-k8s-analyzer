package config

import (
    "log"
    "os"
    "strconv"
    "time"
)

type Config struct {
    HTTPAddr          string
    RedisAddr         string
    LogLevel          string
    IngestBufferSize  int
    AnalyticsWindow   int
    RedisDialTimeout  time.Duration
    RedisReadTimeout  time.Duration
    RedisWriteTimeout time.Duration
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func atoiEnv(key string, def int) int {
    v := getenv(key, "")
    if v == "" {
        return def
    }
    n, err := strconv.Atoi(v)
    if err != nil {
        log.Printf("invalid int for %s: %v, using default %d", key, err, def)
        return def
    }
    return n
}

func durationEnv(key string, def time.Duration) time.Duration {
    v := getenv(key, "")
    if v == "" {
        return def
    }
    d, err := time.ParseDuration(v)
    if err != nil {
        log.Printf("invalid duration for %s: %v, using default %s", key, err, def)
        return def
    }
    return d
}

func Load() Config {
    cfg := Config{
        HTTPAddr:          getenv("HTTP_ADDR", ":8080"),
        RedisAddr:         getenv("REDIS_ADDR", "redis:6379"),
        LogLevel:          getenv("LOG_LEVEL", "info"),
        IngestBufferSize:  atoiEnv("INGEST_BUFFER_SIZE", 1000),
        AnalyticsWindow:   atoiEnv("ANALYTICS_WINDOW_SIZE", 50),
        RedisDialTimeout:  durationEnv("REDIS_DIAL_TIMEOUT", 2*time.Second),
        RedisReadTimeout:  durationEnv("REDIS_READ_TIMEOUT", 500*time.Millisecond),
        RedisWriteTimeout: durationEnv("REDIS_WRITE_TIMEOUT", 500*time.Millisecond),
    }

    if cfg.IngestBufferSize <= 0 {
        cfg.IngestBufferSize = 1000
    }
    if cfg.AnalyticsWindow <= 0 {
        cfg.AnalyticsWindow = 50
    }

    return cfg
}
