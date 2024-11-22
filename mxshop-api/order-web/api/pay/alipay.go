package pay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"

	"mxshop-api/order-web/global"
	"mxshop-api/order-web/proto"
)

func Notify(ctx *gin.Context) {
	//支付宝回调通知
	client, err := alipay.New(global.ServerConfig.AliPayInfo.AppId, global.ServerConfig.AliPayInfo.PrivateKey, false)
	if err != nil {
		zap.S().Errorw("实例化支付宝失败")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	err = client.LoadAliPayPublicKey((global.ServerConfig.AliPayInfo.AliPublicKey))
	if err != nil {
		zap.S().Errorw("加载支付宝的公钥失败")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	//接收alipay回调信息，回调信息有调用支付接口上传的所有信息
	noti, err := client.GetTradeNotification(ctx.Request)
	if err != nil {
		fmt.Println("交易状态为:", noti.TradeStatus)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	//更新订单状态
	_, err = global.OrderSrvClient.UpdateOrderStatus(context.Background(), &proto.OrderStatus{
		OrderSn: noti.OutTradeNo,
		Status:  string(noti.TradeStatus),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	//向alipay回复确认
	ctx.String(http.StatusOK, "success")
}
