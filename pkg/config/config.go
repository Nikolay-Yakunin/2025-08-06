package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              string
	MaxTasks          int8
	MaxFiles          int
	MaxFileSize       int64
	TmpPath           string
	AllowedExtensions []string
	Mode              string
}

// Конструктор конфига
func NewConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", ":8080"),
		MaxTasks:          parseInt8Env("MAX_TASKS", 3),
		MaxFiles:          parseIntEnv("MAX_FILES", 3),
		MaxFileSize:       parseInt64Env("MAX_FILE_SIZE_MB", 300) * 1024 * 1024,
		TmpPath:           getEnv("TMP_PATH", "/tmp/archiver/"),
		AllowedExtensions: strings.Split(getEnv("ALLOWED_EXT", ".jpg .jepg .pdf"), " "),
		Mode:              getEnv("MODE", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 8)
	if err != nil {
		return defaultValue
	}

	return int(value)
}

// Это можно просто с помощью parseIntEnv, но я написал это перевее
func parseInt8Env(key string, defaultValue int8) int8 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 8)
	if err != nil {
		return defaultValue
	}

	return int8(value)
}

func parseInt64Env(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
