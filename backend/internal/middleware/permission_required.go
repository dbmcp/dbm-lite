/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package middleware

import (
	"net/http"
	"strings"

	"dbm-lite/internal/service"

	"github.com/gin-gonic/gin"
)

// ====== 基于权限码的接口级权限中间件 ======

// 接口路径 -> 权限码 映射表
// 接口级权限校验：当请求路径命中时，要求当前登录用户具备对应权限码
var pathPermMap = map[string]string{
	"POST:/api/account":          "account:edit",
	"PUT:/api/account":           "account:edit",
	"DELETE:/api/account":        "account:edit",
	"POST:/api/account/pwd":      "account:reset_pwd",
	"POST:/api/account/lock":     "account:lock",
	"POST:/api/role":             "role:edit",
	"PUT:/api/role":              "role:edit",
	"DELETE:/api/role":           "role:edit",
	"POST:/api/role/perm":        "role:assign_perm",
	"POST:/api/priv/apply/approve": "priv:apply:oper",
	"POST:/api/priv/apply/reject":  "priv:apply:oper",
	"DELETE:/api/priv/priv":      "priv:apply:oper",
	"POST:/api/priv/audit/export": "priv:audit:oper",
}

// PermissionRequired 接口级权限校验中间件
func PermissionRequired(c *gin.Context) {
	// 1. 获取当前用户 ID
	userID := GetStr(c, "userId")
	role := GetStr(c, "role")

	// 2. super_admin 角色直接放行
	if role == "admin" {
		c.Next()
		return
	}

	// 3. 查找当前请求是否需要权限码校验
	key := c.Request.Method + ":" + c.Request.URL.Path
	permCode, matched := pathPermMap[key]
	if !matched {
		// 前缀匹配：/api/account/:id -> 用 /api/account 匹配
		for pattern, code := range pathPermMap {
			parts := strings.SplitN(pattern, ":", 2)
			if len(parts) == 2 && parts[0] == c.Request.Method &&
				strings.HasPrefix(c.Request.URL.Path, parts[1]) {
				permCode = code
				matched = true
				break
			}
		}
	}
	if !matched {
		// 不在权限映射表中 => 默认放行（由业务代码自行判断）
		c.Next()
		return
	}

	// 4. 校验用户是否具备权限码
	codes, err := service.NewAccountService().GetUserPermissionCodes(userID)
	if err != nil {
		Fail(c, http.StatusForbidden, 403, "权限校验失败")
		c.Abort()
		return
	}
	if !codes[permCode] {
		Fail(c, http.StatusForbidden, 403, "缺少接口权限: "+permCode)
		c.Abort()
		return
	}
	c.Next()
}

