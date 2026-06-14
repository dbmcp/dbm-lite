/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package service

import (
	"time"

	"dbm-lite/config"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateTokenLocal(userId, username, role, displayName string) (string, error) {
	claims := jwt.MapClaims{
		"userId":      userId,
		"username":    username,
		"role":        role,
		"displayName": displayName,
		"iss":         "dbm-lite",
		"exp":         time.Now().Add(time.Duration(config.App.TokenTTL) * time.Second).Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App.JWTSecret))
}
