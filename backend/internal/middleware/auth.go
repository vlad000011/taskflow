package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/httpx"
	"taskflow/backend/internal/service"
)

func Auth(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация")
			c.Abort()
			return
		}
		id, role, err := service.ParseJWT(cfg.JWTSecret, parts[1], "access")
		if err != nil {
			httpx.Error(c, 401, "INVALID_TOKEN", "Недействительный токен")
			c.Abort()
			return
		}
		c.Set("userID", id)
		c.Set("role", role)
		c.Next()
	}
}
func Roles(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !allowed[role.(string)] {
			httpx.Error(c, 403, "FORBIDDEN", "Недостаточно прав")
			c.Abort()
			return
		}
		c.Next()
	}
}
