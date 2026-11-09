package middleware

import (
	"net/http"
	"strings"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/httpx"
	"taskflow/backend/internal/model"
	"taskflow/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Auth(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			httpx.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация")
			c.Abort()
			return
		}
		id, _, err := service.ParseJWT(cfg.JWTSecret, parts[1], "access")
		if err != nil {
			httpx.Error(c, 401, "INVALID_TOKEN", "Недействительный токен")
			c.Abort()
			return
		}
		var user model.User
		if db.First(&user, id).Error != nil || user.IsBlocked {
			httpx.Error(c, http.StatusUnauthorized, "ACCOUNT_UNAVAILABLE", "Аккаунт не найден или заблокирован")
			c.Abort()
			return
		}
		c.Set("userID", id)
		c.Set("role", user.Role)
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
		if roleName, ok := role.(string); !ok || !allowed[roleName] {
			httpx.Error(c, 403, "FORBIDDEN", "Недостаточно прав")
			c.Abort()
			return
		}
		c.Next()
	}
}
