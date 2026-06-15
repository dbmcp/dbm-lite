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
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DBType        string
	DBPath        string
	MySQLHost     string
	MySQLPort     int
	MySQLUsername string
	MySQLPassword string
	MySQLDatabase string
	JWTSecret     string
	AESKey        string
	AdminUsername string
	AdminPassword string
	TokenTTL      int
}

var App Config

func Load() error {
	_ = godotenv.Load()
	dbType := strings.ToLower(strings.TrimSpace(getEnv("DBM_LITE_DB_TYPE", "sqlite")))
	if dbType != "mysql" {
		dbType = "sqlite"
	}
	App = Config{
		ServerPort:    getEnv("DBM_LITE_SERVER_PORT", "8080"),
		DBType:        dbType,
		DBPath:        getEnv("DBM_LITE_DB_PATH", "./data/dbm-lite.db"),
		MySQLHost:     getEnv("DBM_LITE_MYSQL_HOST", "127.0.0.1"),
		MySQLPort:     getEnvInt("DBM_LITE_MYSQL_PORT", 3306),
		MySQLUsername: getEnv("DBM_LITE_MYSQL_USERNAME", "root"),
		MySQLPassword: getEnv("DBM_LITE_MYSQL_PASSWORD", ""),
		MySQLDatabase: getEnv("DBM_LITE_MYSQL_DATABASE", "dbm_lite"),
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
