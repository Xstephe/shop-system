package initialize

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mxshop-api/user-web/middlewares"
	router2 "mxshop-api/user-web/router"
)

func Routers() *gin.Engine {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})
	//配置跨域
	router.Use(middlewares.Cors())
	ApiGroup := router.Group("/u/v1")
	router2.InitUserRouter(ApiGroup)
	router2.InitBaseRouter(ApiGroup)
	return router
}
