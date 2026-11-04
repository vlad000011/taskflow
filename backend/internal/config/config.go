package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL, JWTSecret, CORSOrigin, UploadDir string
	Port                                          string
	AccessMinutes, RefreshDays, MaxUploadMB       int
	SeedDemo                                      bool
}

func Load() Config {
	return Config{DatabaseURL: env("DATABASE_URL", "postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable"), JWTSecret: env("JWT_SECRET", "local-development-secret-change-this-32chars"), CORSOrigin: env("CORS_ORIGIN", "http://localhost:5173"), UploadDir: env("UPLOAD_DIR", "./uploads"), Port: env("API_PORT", "8081"), AccessMinutes: envInt("ACCESS_TOKEN_MINUTES", 15), RefreshDays: envInt("REFRESH_TOKEN_DAYS", 14), MaxUploadMB: envInt("MAX_UPLOAD_MB", 5), SeedDemo: os.Getenv("SEED_DEMO") == "true"}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func envInt(k string, d int) int {
	v, e := strconv.Atoi(os.Getenv(k))
	if e != nil || v < 1 {
		return d
	}
	return v
}
