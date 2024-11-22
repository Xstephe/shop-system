package goods

import (
	"context"
	"mxshop-api/goods-web/api"
	"mxshop-api/goods-web/forms"
	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/proto"
	"net/http"
	"strconv"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 商品列表
func List(ctx *gin.Context) {
	request := proto.GoodsFilterRequest{}

	priceMin := ctx.DefaultQuery("pmin", "0")
	priceMinInt, _ := strconv.Atoi(priceMin)
	request.PriceMin = int32(priceMinInt)

	priceMax := ctx.DefaultQuery("pmax", "0")
	priceMaxInt, _ := strconv.Atoi(priceMax)
	request.PriceMax = int32(priceMaxInt)

	isHot := ctx.DefaultQuery("ih", "0")
	if isHot == "1" {
		request.IsHot = true
	}

	isNew := ctx.DefaultQuery("in", "0")
	if isNew == "1" {
		request.IsNew = true
	}

	isTab := ctx.DefaultQuery("it", "0")
	if isTab == "1" {
		request.IsTab = true
	}

	categoryId := ctx.DefaultQuery("c", "0")
	categoryIdInt, _ := strconv.Atoi(categoryId)
	request.TopCategory = int32(categoryIdInt)

	pages := ctx.DefaultQuery("pn", "0")
	pagesInt, _ := strconv.Atoi(pages)
	request.Pages = int32(pagesInt)

	perNums := ctx.DefaultQuery("pnum", "0")
	perNumsInt, _ := strconv.Atoi(perNums)
	request.PagePerNums = int32(perNumsInt)

	keyWords := ctx.DefaultQuery("q", "")
	request.KeyWords = keyWords

	brandId := ctx.DefaultQuery("b", "0")
	brandIdInt, _ := strconv.Atoi(brandId)
	request.Brand = int32(brandIdInt)

	//请求商品的service服务   加上链路追踪  加上限流熔断
	e, b := sentinel.Entry("goods-list", sentinel.WithTrafficType(base.Inbound))
	if b != nil {
		ctx.JSON(http.StatusTooManyRequests, gin.H{
			"msg": "请求过于频繁，请稍后重试",
		})
		return
	}

	r, err := global.GoodsSrvClient.GoodsList(context.WithValue(context.Background(), "ginContext", ctx), &request)
	if err != nil {
		zap.S().Errorw("[List] 查询【商品列表】失败")
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	e.Exit()

	var goodsList = make([]interface{}, 0)
	for _, value := range r.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc, // 设置的时候grpc是和name一样的
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"category": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":  value.IsHot,
			"is_new":  value.IsNew,
			"on_sale": value.OnSale,
		})
	}
	var rspMap = make(map[string]interface{})
	rspMap["total"] = r.Total
	rspMap["data"] = goodsList

	ctx.JSON(http.StatusOK, rspMap)
}

// 新建商品
func New(ctx *gin.Context) {
	var goodsform forms.GoodsForm
	if err := ctx.ShouldBindJSON(&goodsform); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}
	goodsInfoResponse, err := global.GoodsSrvClient.CreateGoods(context.WithValue(context.Background(), "ginContext", ctx), &proto.CreateGoodsInfo{
		Name:            goodsform.Name,
		GoodsSn:         goodsform.GoodsSn,
		Stocks:          goodsform.Stocks,
		MarketPrice:     goodsform.MarketPrice,
		ShopPrice:       goodsform.ShopPrice,
		GoodsBrief:      goodsform.GoodsBrief,
		ShipFree:        *goodsform.ShipFree,
		Images:          goodsform.Images,
		DescImages:      goodsform.DescImages,
		GoodsFrontImage: goodsform.FrontImage,
		CategoryId:      goodsform.CategoryId,
		BrandId:         goodsform.Brand,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, goodsInfoResponse)
}

// 根据商品Id获取商品详情
func Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	goodsInfoResponse, err := global.GoodsSrvClient.GetGoodsDetail(context.WithValue(context.Background(), "ginContext", ctx), &proto.GoodInfoRequest{
		Id: int32(i),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	RspGoods := map[string]interface{}{
		"id":          goodsInfoResponse.Id,
		"name":        goodsInfoResponse.Name,
		"goods_brief": goodsInfoResponse.GoodsBrief,
		"desc":        goodsInfoResponse.GoodsDesc,
		"ship_free":   goodsInfoResponse.ShipFree,
		"images":      goodsInfoResponse.Images,
		"desc_images": goodsInfoResponse.DescImages,
		"front_image": goodsInfoResponse.GoodsFrontImage,
		"shop_price":  goodsInfoResponse.ShopPrice,
		"category": map[string]interface{}{
			"id":   goodsInfoResponse.Category.Id,
			"name": goodsInfoResponse.Category.Name,
		},
		"brand": map[string]interface{}{
			"id":   goodsInfoResponse.Brand.Id,
			"name": goodsInfoResponse.Brand.Name,
			"logo": goodsInfoResponse.Brand.Logo,
		},
		"is_hot":  goodsInfoResponse.IsHot,
		"is_new":  goodsInfoResponse.IsNew,
		"on_sale": goodsInfoResponse.OnSale,
	}

	ctx.JSON(http.StatusOK, RspGoods)
}

// 根据商品Id删除商品
func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteGoods(context.WithValue(context.Background(), "ginContext", ctx), &proto.DeleteGoodsInfo{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

// 获取商品库存
func Stocks(ctx *gin.Context) {
	id := ctx.Param("id")
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	//TODO 商品的库存
}

// 更新商品
func Update(ctx *gin.Context) {
	GoodsFrom := forms.GoodsForm{}
	if err := ctx.ShouldBind(&GoodsFrom); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	goodsId := ctx.Param("id")
	i, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateGoods(context.Background(), &proto.CreateGoodsInfo{
		Id:              int32(i),
		Name:            GoodsFrom.Name,
		GoodsSn:         GoodsFrom.GoodsSn,
		Stocks:          GoodsFrom.Stocks,
		MarketPrice:     GoodsFrom.MarketPrice,
		ShopPrice:       GoodsFrom.ShopPrice,
		GoodsBrief:      GoodsFrom.GoodsBrief,
		ShipFree:        *GoodsFrom.ShipFree,
		Images:          GoodsFrom.Images,
		DescImages:      GoodsFrom.DescImages,
		GoodsFrontImage: GoodsFrom.FrontImage,
		CategoryId:      GoodsFrom.CategoryId,
		BrandId:         GoodsFrom.Brand,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}

// 更新商品状态
func UpdateStatus(ctx *gin.Context) {
	//获取表单数据
	goodsStatusForm := forms.GoodsStatusForm{}
	if err := ctx.ShouldBind(&goodsStatusForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	goodsId := ctx.Param("id")
	id, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	//获取商品对应的品牌和分类
	rsp, err := global.GoodsSrvClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: int32(id),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
	}

	createGoods := proto.CreateGoodsInfo{
		Id:              int32(id),
		CategoryId:      rsp.CategoryId,
		BrandId:         rsp.Brand.Id,
		Name:            rsp.Name,
		GoodsSn:         rsp.GoodsSn,
		MarketPrice:     rsp.MarketPrice,
		ShopPrice:       rsp.ShopPrice,
		GoodsBrief:      rsp.GoodsBrief,
		ShipFree:        rsp.ShipFree,
		Images:          rsp.Images,
		DescImages:      rsp.DescImages,
		GoodsFrontImage: rsp.GoodsFrontImage,
		IsNew:           *goodsStatusForm.IsNew,
		IsHot:           *goodsStatusForm.IsHot,
		OnSale:          *goodsStatusForm.OnSale,
	}

	_, err = global.GoodsSrvClient.UpdateGoods(context.Background(), &createGoods)
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "修改成功",
	})
}
