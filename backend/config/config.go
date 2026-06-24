package config

import (
"fmt"
"os"
"strconv"
"strings"

"github.com/joho/godotenv"
)

type Config struct {
ServerPort    string
JWTSecret     string
TokenTTL      int
DBType        string
DBPath        string
MySQLUsername string
MySQLPassword string
MySQLHost     string
MySQLPort     string
MySQLDatabase string
AdminUsername string
AdminPassword string
AESKey        string
APITimeout    int // API请求超时时间（秒）
}

var App Config

func getEnv(key, defaultValue string) string {
if value, exists := os.LookupEnv(key); exists {
return value
}
return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
if value, exists := os.LookupEnv(key); exists {
if intValue, err := strconv.Atoi(value); err == nil {
return intValue
}
}
return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
if value, exists := os.LookupEnv(key); exists {
v := strings.ToLower(strings.TrimSpace(value))
if v == "true" || v == "1" || v == "yes" || v == "on" {
return true
}
if v == "false" || v == "0" || v == "no" || v == "off" {
return false
}
}
return defaultValue
}

func Load() error {
if err := godotenv.Load(); err != nil {
fmt.Printf("[config] Warning: .env file not found, using defaults: %v\n", err)
}

App = Config{
ServerPort:    getEnv("DBM_LITE_SERVER_PORT", "8080"),
JWTSecret:     getEnv("DBM_LITE_JWT_SECRET", "dbm-lite-jwt-secret-key-change-in-production"),
TokenTTL:      getEnvInt("DBM_LITE_TOKEN_TTL_SECONDS", 86400),
DBType:        getEnv("DBM_LITE_DB_TYPE", "mysql"),
DBPath:        getEnv("DBM_LITE_DB_PATH", "./data/dbm-lite.db"),
MySQLUsername: getEnv("DBM_LITE_MYSQL_USERNAME", "root"),
MySQLPassword: getEnv("DBM_LITE_MYSQL_PASSWORD", "root"),
MySQLHost:     getEnv("DBM_LITE_MYSQL_HOST", "127.0.0.1"),
MySQLPort:     getEnv("DBM_LITE_MYSQL_PORT", "3306"),
MySQLDatabase: getEnv("DBM_LITE_MYSQL_DATABASE", "dbm_lite"),
AdminUsername: getEnv("DBM_LITE_ADMIN_USERNAME", "admin"),
AdminPassword: getEnv("DBM_LITE_ADMIN_PASSWORD", "admin123"),
AESKey:        getEnv("DBM_LITE_AES_KEY", "dbm-lite-aes-key-32bytes-long-please-"),
APITimeout:    getEnvInt("DBM_LITE_API_TIMEOUT", 60),
}

return nil
}