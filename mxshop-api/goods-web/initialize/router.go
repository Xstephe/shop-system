package initialize

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mxshop-api/goods-web/middlewares"
	router2 "mxshop-api/goods-web/router"
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
	ApiGroup := router.Group("/v1")
	router2.InitGoodsRouter(ApiGroup)
	router2.InitCategory(ApiGroup)
	router2.InitBanner(ApiGroup)
	router2.InitBrands(ApiGroup)
	return router
}
