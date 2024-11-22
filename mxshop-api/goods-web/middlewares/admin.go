package middlewares

import (
	"mxshop-api/goods-web/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 访问权限配置，管理员才能登录
func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := ctx.Get("claims")
		currentUser := claims.(*models.CustomClaims)
		if currentUser.AuthorityId != 2 {
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "无权限",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
