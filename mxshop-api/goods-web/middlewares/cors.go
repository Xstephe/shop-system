package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 为option请求做返回，解决跨域问题
func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		ctx.Header("Access-Control-Allow-Headers", "Token, X-CSRT-Token, Content-Type, AccessToken, Authorization,x-token")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}
	}
}
