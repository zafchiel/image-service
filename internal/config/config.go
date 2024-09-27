package config

import "os"

type Config struct {
	DBPath        string
	StoragePath   string
	ServerAddress string
	MaxUploadSize int64
}

func Load() *Config {
	return &Config{
		DBPath:        getEnv("DB_PATH", "sqlite.db"),
		StoragePath:   getEnv("STORAGE_PATH", "assets"),
		ServerAddress: getEnv("PORT", ":8080"),
		MaxUploadSize: 10 << 20, // 10 MB
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
