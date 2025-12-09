package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/example/go-k8s-analyzer/internal/config"
    "github.com/example/go-k8s-analyzer/internal/httpserver"
    "github.com/example/go-k8s-analyzer/internal/logger"
)

func main() {
    cfg := config.Load()
    logg := logger.New(cfg.LogLevel)

    srv := httpserver.New(cfg, logg)

    go func() {
        if err := srv.Start(); err != nil {
            log.Fatalf("server error: %v", err)
        }
    }()

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
    <-stopCh

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Stop(ctx)
}
