package router

import (
	"github.com/gin-gonic/gin"

	"mxshop-api/goods-web/api/banner"
	"mxshop-api/goods-web/middlewares"
)

func InitBanner(router *gin.RouterGroup) {
	BannerRouter := router.Group("banners").Use(middlewares.Trace())
	{
		BannerRouter.GET("", banner.List)                                                            //获取轮播图列表
		BannerRouter.POST("", middlewares.JWTAuth(), middlewares.IsAdminAuth(), banner.New)          //添加轮播图
		BannerRouter.DELETE("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), banner.Delete) //删除轮播图
		BannerRouter.PUT("/:id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), banner.Update)    //更新轮播图
	}
}
