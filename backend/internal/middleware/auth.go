/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DBA老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package middleware

import (
	"net/http"
	"strings"
	"time"

	"dbm-lite/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 0
	CodeParam   = 400
	CodeAuth    = 401
	CodeForbid  = 403
	CodeServer  = 500
)

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Message: "成功", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.AbortWithStatusJSON(httpStatus, Response{Code: code, Message: message})
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			Fail(c, http.StatusUnauthorized, CodeAuth, "未登录或token无效")
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.App.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			Fail(c, http.StatusUnauthorized, CodeAuth, "token无效或已过期")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			Fail(c, http.StatusUnauthorized, CodeAuth, "token无效")
			return
		}
		userId, _ := claims["userId"].(string)
		username, _ := claims["username"].(string)
		role, _ := claims["role"].(string)
		displayName, _ := claims["displayName"].(string)
		c.Set("userId", userId)
		c.Set("username", username)
		c.Set("role", role)
		c.Set("displayName", displayName)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			Fail(c, http.StatusForbidden, CodeForbid, "需要管理员权限")
			return
		}
		c.Next()
	}
}

func GenerateToken(userId, username, role, displayName string) (string, error) {
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

func GetStr(c *gin.Context, key string) string {
	v, _ := c.Get(key)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func GetClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	return ip
}

