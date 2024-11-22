package order

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"

	"mxshop-api/order-web/api"
	"mxshop-api/order-web/forms"
	"mxshop-api/order-web/global"
	"mxshop-api/order-web/models"
	"mxshop-api/order-web/proto"
)

// 订单列表
func List(ctx *gin.Context) {
	userId, _ := ctx.Get("userId")
	claims, _ := ctx.Get("claims")

	request := proto.OrderFilterRequest{}
	model := claims.(*models.CustomClaims)
	//如果是管理员用户，则返回所有的订单
	if model.AuthorityId == 1 {
		request.UserId = int32(userId.(uint))
	}

	//分页
	Pages := ctx.DefaultQuery("p", "0")
	PageInt, _ := strconv.Atoi(Pages)
	request.Pages = int32(PageInt)

	PageNums := ctx.DefaultQuery("pnum", "0")
	PageNumsInt, _ := strconv.Atoi(PageNums)
	request.PagePerNums = int32(PageNumsInt)

	rsp, err := global.OrderSrvClient.OrderList(context.Background(), &request)
	if err != nil {
		zap.S().Errorw("获取订单列表失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	OrderList := make([]interface{}, 0)
	for _, Item := range rsp.Data {
		ItemMap := map[string]interface{}{}

		ItemMap["id"] = Item.Id
		ItemMap["name"] = Item.Name
		ItemMap["total"] = Item.Total
		ItemMap["userId"] = Item.UserId
		ItemMap["status"] = Item.Status
		ItemMap["order_sn"] = Item.OrderSn
		ItemMap["address"] = Item.Address
		ItemMap["mobile"] = Item.Mobile
		ItemMap["post"] = Item.Post
		ItemMap["pay_type"] = Item.PayType
		ItemMap["add_time"] = Item.AddTime
		OrderList = append(OrderList, Item)
	}

	ReMap := gin.H{
		"total": rsp.Total,
		"data":  OrderList,
	}
	ctx.JSON(http.StatusOK, ReMap)

}

// 新建订单
func CreatOrder(ctx *gin.Context) {
	var OrderForm forms.OrderForms
	if err := ctx.ShouldBindJSON(&OrderForm); err != nil {
		api.HandleValidatorError(err, ctx)
		return
	}

	userId, _ := ctx.Get("userId")

	Rsp, err := global.OrderSrvClient.CreateOrder(context.WithValue(context.Background(), "ginContext", ctx), &proto.OrderRequest{
		UserId:  int32(userId.(uint)),
		Name:    OrderForm.Name,
		Address: OrderForm.Address,
		Mobile:  OrderForm.Mobile,
		Post:    OrderForm.Post,
	})
	if err != nil {
		zap.S().Errorw("新建订单失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	//TODO 此时的逻辑跳转至支付宝支付页面，可通过web层或是srv层返回支付宝支付URL
	// 生成支付宝的url
	client, err := alipay.New(global.ServerConfig.AliPayInfo.AppId, global.ServerConfig.AliPayInfo.PrivateKey, false)
	if err != nil {
		zap.S().Errorw("初始化支付宝支付对象失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	err = client.LoadAliPayPublicKey(global.ServerConfig.AliPayInfo.AliPublicKey)
	if err != nil {
		zap.S().Errorw("加载alipay公钥失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	p := alipay.TradePagePay{}
	p.NotifyURL = global.ServerConfig.AliPayInfo.NotifyUrl
	p.ReturnURL = global.ServerConfig.AliPayInfo.ReturnUrl
	p.Subject = "mxshop订单-" + Rsp.OrderSn
	p.OutTradeNo = Rsp.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(Rsp.Total), 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err := client.TradePagePay(p)
	if err != nil {
		zap.S().Errorw("生成alipay支付url失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":         Rsp.Id,
		"alipay_url": url.String(),
	})
}

// 获取订单详情
func DetailOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "Url格式出错",
		})
		return
	}

	request := proto.OrderRequest{
		Id: int32(i),
	}

	userId, _ := ctx.Get("userId")
	claims, _ := ctx.Get("claims")
	model := claims.(*models.CustomClaims)
	//如果是管理员用户，则返回所有的订单
	if model.AuthorityId == 1 {
		request.UserId = int32(userId.(uint))
	}

	Rsp, err := global.OrderSrvClient.OrderDetail(context.Background(), &request)
	if err != nil {
		zap.S().Errorw("获取订单列表失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	reMap := gin.H{}
	reMap["id"] = Rsp.OrderInfo.Id
	reMap["orderSn"] = Rsp.OrderInfo.OrderSn
	reMap["name"] = Rsp.OrderInfo.Name
	reMap["userId"] = Rsp.OrderInfo.UserId
	reMap["status"] = Rsp.OrderInfo.Status
	reMap["payType"] = Rsp.OrderInfo.PayType
	reMap["post"] = Rsp.OrderInfo.Post
	reMap["address"] = Rsp.OrderInfo.Address
	reMap["mobile"] = Rsp.OrderInfo.Mobile
	reMap["total"] = Rsp.OrderInfo.Total
	reMap["addTime"] = Rsp.OrderInfo.AddTime

	GoodsList := make([]interface{}, 0)
	for _, goods := range Rsp.Goods {
		goodsItem := map[string]interface{}{}
		goodsItem["id"] = goods.Id
		goodsItem["name"] = goods.GoodsName
		goodsItem["goodsId"] = goods.GoodsId
		goodsItem["image"] = goods.GoodsImage
		goodsItem["nums"] = goods.Nums
		goodsItem["price"] = goods.GoodsPrice
		goodsItem["orderId"] = goods.OrderId
		GoodsList = append(GoodsList, goodsItem)
	}
	reMap["goods"] = GoodsList

	// 生成支付宝的url
	client, err := alipay.New(global.ServerConfig.AliPayInfo.AppId, global.ServerConfig.AliPayInfo.PrivateKey, false)
	if err != nil {
		zap.S().Errorw("初始化支付宝支付对象失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	err = client.LoadAliPayPublicKey(global.ServerConfig.AliPayInfo.AliPublicKey)
	if err != nil {
		zap.S().Errorw("加载alipay公钥失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	p := alipay.TradePagePay{}
	p.NotifyURL = global.ServerConfig.AliPayInfo.NotifyUrl
	p.ReturnURL = global.ServerConfig.AliPayInfo.ReturnUrl
	p.Subject = "mxshop订单-" + Rsp.OrderInfo.OrderSn
	p.OutTradeNo = Rsp.OrderInfo.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(Rsp.OrderInfo.Total), 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err := client.TradePagePay(p)
	if err != nil {
		zap.S().Errorw("生成alipay支付url失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	reMap["alipay_url"] = url.String()
	ctx.JSON(http.StatusOK, reMap)

}
