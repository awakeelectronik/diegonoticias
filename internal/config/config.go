package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string
	Listen   string
	DataDir  string
	UploadsDir string
	SiteDir  string
	HugoBin  string
	PagefindBin string
	LogLevel slog.Level
}

func Load() (Config, error) {
	env := strings.TrimSpace(os.Getenv("DN_ENV"))
	if env == "" {
		env = "development"
	}

	if env == "development" {
		_ = godotenv.Load()
	}

	listen := strings.TrimSpace(os.Getenv("DN_LISTEN"))
	if listen == "" {
		listen = "127.0.0.1:8080"
	}
	dataDir := strings.TrimSpace(os.Getenv("DN_DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data"
	}
	uploadsDir := strings.TrimSpace(os.Getenv("DN_UPLOADS_DIR"))
	if uploadsDir == "" {
		uploadsDir = "./static-uploads"
	}
	siteDir := strings.TrimSpace(os.Getenv("DN_SITE_DIR"))
	if siteDir == "" {
		siteDir = "./site"
	}
	hugoBin := strings.TrimSpace(os.Getenv("DN_HUGO_BIN"))
	if hugoBin == "" {
		hugoBin = "hugo"
	}
	pagefindBin := strings.TrimSpace(os.Getenv("DN_PAGEFIND_BIN"))
	if pagefindBin == "" {
		pagefindBin = "pagefind"
	}

	levelStr := strings.TrimSpace(os.Getenv("DN_LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info"
	}

	level, err := parseLevel(levelStr)
	if err != nil {
		return Config{}, fmt.Errorf("DN_LOG_LEVEL: %w", err)
	}

	return Config{
		Env:      env,
		Listen:   listen,
		DataDir:  dataDir,
		UploadsDir: uploadsDir,
		SiteDir:  siteDir,
		HugoBin:  hugoBin,
		PagefindBin: pagefindBin,
		LogLevel: level,
	}, nil
}

func parseLevel(s string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, errors.New("debe ser debug|info|warn|error")
	}
}

