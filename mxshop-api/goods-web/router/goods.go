package router

import (
	"github.com/gin-gonic/gin"

	"mxshop-api/goods-web/api/goods"
	"mxshop-api/goods-web/middlewares"
)

func InitGoodsRouter(Router *gin.RouterGroup) {
	GoodsRouter := Router.Group("goods").Use(middlewares.Trace())
	{
		GoodsRouter.GET("", goods.List)                                                                //获取商品列表
		GoodsRouter.POST("", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.New)              //需要用户权限，也需要登录 新建商品
		GoodsRouter.GET(":id", goods.Detail)                                                           //获取商品详情
		GoodsRouter.DELETE(":id", goods.Delete)                                                        //删除商品
		GoodsRouter.GET(":id/stocks", goods.Stocks)                                                    //获取商品的库存
		GoodsRouter.PUT(":id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.Update)         //更新商品
		GoodsRouter.PATCH(":id", middlewares.JWTAuth(), middlewares.IsAdminAuth(), goods.UpdateStatus) //更新状态
	}
}
