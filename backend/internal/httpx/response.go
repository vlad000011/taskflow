package httpx

import "github.com/gin-gonic/gin"

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
func OK(c *gin.Context, data any) { c.JSON(200, gin.H{"data": data}) }
