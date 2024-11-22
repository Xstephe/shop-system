package initialize

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mxshop-api/userop-web/middlewares"
	router2 "mxshop-api/userop-web/router"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})

	//配置跨域
	Router.Use(middlewares.Cors())
	ApiGroup := Router.Group("/up/v1")
	router2.InitUserFavRouter(ApiGroup)
	router2.InitMessageRouter(ApiGroup)
	router2.InitAddressRouter(ApiGroup)
	return Router
}
