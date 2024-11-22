package router

import (
	"github.com/gin-gonic/gin"

	"mxshop-api/order-web/api/order"
	"mxshop-api/order-web/api/pay"
	"mxshop-api/order-web/middlewares"
)

func InitOrderRouter(router *gin.RouterGroup) {
	OrderRouter := router.Group("order").Use(middlewares.JWTAuth()).Use(middlewares.Trace())
	{
		//中间件的参数位置需要按业务需求而定
		OrderRouter.GET("", order.List)            //订单列表
		OrderRouter.GET("/:id", order.DetailOrder) //获取订单详细
		OrderRouter.POST("", order.CreatOrder)     //新建订单
	}

	PayRouter := router.Group("pay")
	{
		PayRouter.POST("alipay/notify", pay.Notify)
	}
}
