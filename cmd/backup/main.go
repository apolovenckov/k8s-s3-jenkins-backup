package main

import (
    "log"
    "os"
    "strings"

    "k8s-s3-backup/internal/backup"
    "k8s-s3-backup/internal/config"
)

func main() {
    log.SetFlags(0)
    cfg, err := config.FromEnv(os.Environ())
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    if err := backup.Run(cfg); err != nil {
        log.Fatalf("backup failed: %s", strings.TrimSpace(err.Error()))
    }
}
