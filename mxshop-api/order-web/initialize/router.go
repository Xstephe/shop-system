package initialize

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mxshop-api/order-web/middlewares"
	router2 "mxshop-api/order-web/router"
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
	ApiGroup := Router.Group("/o/v1")
	router2.InitOrderRouter(ApiGroup)
	router2.InitShopCartRouter(ApiGroup)
	return Router
}
