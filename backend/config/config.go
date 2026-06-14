/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DBPath        string
	JWTSecret     string
	AESKey        string
	AdminUsername string
	AdminPassword string
	TokenTTL      int
}

var App Config

func Load() error {
	_ = godotenv.Load()
	App = Config{
		ServerPort:    getEnv("DBM_LITE_SERVER_PORT", "8080"),
		DBPath:        getEnv("DBM_LITE_DB_PATH", "./data/dbm-lite.db"),
		JWTSecret:     getEnv("DBM_LITE_JWT_SECRET", "dbm-lite-default-jwt-secret-change-me"),
		AESKey:        getEnv("DBM_LITE_AES_KEY", "dbm-lite-aes-key-change-me-32-bytes!!"),
		AdminUsername: getEnv("DBM_LITE_ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("DBM_LITE_ADMIN_PASSWORD", "admin123"),
		TokenTTL:      getEnvInt("DBM_LITE_TOKEN_TTL_SECONDS", 86400),
	}
	if err := os.MkdirAll("./data", 0755); err != nil {
		// 忽略错误，可能已经存在
	}
	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
