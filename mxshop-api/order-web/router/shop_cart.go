package router

import (
	"github.com/gin-gonic/gin"

	"mxshop-api/order-web/api/shop_cart"
	"mxshop-api/order-web/middlewares"
)

func InitShopCartRouter(router *gin.RouterGroup) {
	ShopCartRouter := router.Group("shopcarts").Use(middlewares.JWTAuth())
	{
		ShopCartRouter.GET("", shop_cart.List)                  //获取购物车列表
		ShopCartRouter.POST("", shop_cart.CreateCartItem)       //加入购物车
		ShopCartRouter.PATCH("/:id", shop_cart.UpdateCartItem)  //更新购物车
		ShopCartRouter.DELETE("/:id", shop_cart.DeleteCartItem) //移除购物车
	}
}
