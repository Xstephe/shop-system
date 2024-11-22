package shop_cart

import (
	"context"
	"mxshop-api/order-web/api"
	"mxshop-api/order-web/forms"
	"mxshop-api/order-web/global"
	"mxshop-api/order-web/proto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 用户获取购物车列表
func List(ctx *gin.Context) {
	userId, _ := ctx.Get("userId")
	rsp, err := global.OrderSrvClient.CartItemList(context.Background(), &proto.UserInfo{
		Id: int32(userId.(uint)),
	})
	if err != nil {
		zap.S().Errorw("[List] 查询 【购物车列表】失败")
		api.HandleValidatorError(err, ctx)
		return
	}
	ids := make([]int32, 0)
	for _, item := range rsp.Data {
		ids = append(ids, item.GoodsId)
	}
	if len(ids) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"total": 0,
		})
		return
	}

	//请求商品服务，获取商品信息
	goodsRsp, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: ids,
	})
	if err != nil {
		zap.S().Errorw("[List] 批量查询【商品列表】失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	reMap := gin.H{
		"total": rsp.Total,
	}

	goodsList := make([]interface{}, 0)
	for _, item := range rsp.Data {
		for _, good := range goodsRsp.Data {
			if good.Id == item.GoodsId {
				tmpMap := make(map[string]interface{})
				tmpMap["id"] = item.Id
				tmpMap["goods_id"] = item.GoodsId
				tmpMap["good_name"] = good.Name
				tmpMap["good_image"] = good.GoodsFrontImage
				tmpMap["good_price"] = good.ShopPrice
				tmpMap["nums"] = item.Nums
				tmpMap["checked"] = item.Checked
				goodsList = append(goodsList, tmpMap)
			}
		}
	}
	reMap["data"] = goodsList
	ctx.JSON(http.StatusOK, reMap)

}

// 添加商品到购物车
func CreateCartItem(ctx *gin.Context) {
	var cartItem forms.ShopCartForm
	if err := ctx.ShouldBindJSON(&cartItem); err != nil {
		zap.S().Errorw("获取表单失败")
		api.HandleValidatorError(err, ctx)
		return
	}

	//查询商品是否存在
	_, err := global.GoodsSrvClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: cartItem.GoodsId,
	})
	if err != nil {
		zap.S().Errorw("查询【商品信息】失败")
		api.HandleValidatorError(err, ctx)
		return
	}

	//查询库存
	InvGoods, err := global.InventorySrvClient.InvDetail(context.Background(), &proto.GoodsInvInfo{
		GoodsId: cartItem.GoodsId,
	})
	if err != nil {
		zap.S().Errorw("查询【库存信息】失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	//判断库存是否充足
	if cartItem.Nums > InvGoods.Num {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "库存不足",
		})
		return
	}

	userId, _ := ctx.Get("userId")
	CartItemRsp, err := global.OrderSrvClient.CreateCarItem(context.Background(), &proto.CartItemRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: cartItem.GoodsId,
		Nums:    cartItem.Nums,
	})
	if err != nil {
		zap.S().Errorw("加入购物车失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id": CartItemRsp.Id,
	})
}

// 更新购物车状态
func UpdateCartItem(ctx *gin.Context) {
	ShopCartUpdateForm := forms.UpdateShopCartForm{}
	if err := ctx.ShouldBindJSON(&ShopCartUpdateForm); err != nil {
		api.HandleValidatorError(err, ctx)
		return
	}

	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "Url格式出错",
		})
		return
	}

	userId, _ := ctx.Get("userId")
	request := proto.CartItemRequest{
		GoodsId: int32(i),
		UserId:  int32(userId.(uint)),
		Nums:    ShopCartUpdateForm.Nums,
		Checked: false,
	}
	if ShopCartUpdateForm.Checked != nil {
		request.Checked = *ShopCartUpdateForm.Checked
	}

	_, err = global.OrderSrvClient.UpdateCartItem(context.Background(), &request)
	if err != nil {
		zap.S().Errorw("更新购物车记录失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.Status(http.StatusOK)
}

// 移除购物车
func DeleteCartItem(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "Url格式出错",
		})
		return
	}
	userId, _ := ctx.Get("userId")
	_, err = global.OrderSrvClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: int32(i),
	})
	if err != nil {
		zap.S().Errorw("删除购物车记录失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}
