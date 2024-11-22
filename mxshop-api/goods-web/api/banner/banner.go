package banner

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"

	"mxshop-api/goods-web/api"
	"mxshop-api/goods-web/forms"
	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/proto"
)

// BannerList 获取轮播图列表
func List(ctx *gin.Context) {
	Rsp, err := global.GoodsSrvClient.BannerList(context.Background(), &empty.Empty{})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	data := make([]interface{}, 0)
	for _, value := range Rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["index"] = value.Index
		reMap["image"] = value.Image
		reMap["url"] = value.Url

		data = append(data, reMap)
	}
	ctx.JSON(http.StatusOK, data)
}

// NewBanner 添加轮播图
func New(ctx *gin.Context) {
	BannerForm := forms.BannerForm{}
	if err := ctx.ShouldBindJSON(&BannerForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	Rsp, err := global.GoodsSrvClient.CreateBanner(context.Background(), &proto.BannerRequest{
		Index: int32(BannerForm.Index),
		Image: BannerForm.Image,
		Url:   BannerForm.Url,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	response := make(map[string]interface{})
	response["id"] = Rsp.Id
	response["index"] = BannerForm.Index
	response["url"] = BannerForm.Url
	response["image"] = BannerForm.Image

	ctx.JSON(http.StatusOK, response)
}

// DeleteBanner 删除轮播图
func Delete(ctx *gin.Context) {
	bannerId := ctx.Param("id")
	id, err := strconv.ParseInt(bannerId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteBanner(context.Background(), &proto.BannerRequest{
		Id: int32(id),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

// UpdateBanner 更新轮播图
func Update(ctx *gin.Context) {
	updateBanner := forms.BannerForm{}
	if err := ctx.ShouldBindJSON(&updateBanner); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	bannerId := ctx.Param("id")
	id, err := strconv.ParseInt(bannerId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateBanner(context.Background(), &proto.BannerRequest{
		Id:    int32(id),
		Index: int32(updateBanner.Index),
		Image: updateBanner.Image,
		Url:   updateBanner.Url,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "修改成功",
	})
}
