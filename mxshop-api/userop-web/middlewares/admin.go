package middlewares

import (
	"mxshop-api/userop-web/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func IsAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get("claims")
		currentUser := claims.(*models.CustomClaims)
		if currentUser.AuthorityId != 2 {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "无权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
