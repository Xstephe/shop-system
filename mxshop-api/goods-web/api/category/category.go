package category

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"

	"mxshop-api/goods-web/api"
	"mxshop-api/goods-web/forms"
	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/proto"
)

func List(ctx *gin.Context) {
	Rsp, err := global.GoodsSrvClient.GetAllCategorysList(context.Background(), &empty.Empty{})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	data := make([]interface{}, 0)

	//json反序列化
	err = json.Unmarshal([]byte(Rsp.JsonData), &data)
	if err != nil {
		zap.S().Info("List [查询] 【分类列表】失败", err.Error())
	}
	ctx.JSON(http.StatusOK, data)
}

func Detail(ctx *gin.Context) {
	goodsId := ctx.Param("id")
	id, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	reMap := make(map[string]interface{})
	subCategorys := make([]interface{}, 0)
	Rsp, err := global.GoodsSrvClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: int32(id),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	} else {
		for _, value := range Rsp.SubCategorys {
			subCategorys = append(subCategorys, map[string]interface{}{
				"id":              value.Id,
				"name":            value.Name,
				"level":           value.Level,
				"parent_category": value.ParentCategory,
				"is_tab":          value.IsTab,
			})
		}
		reMap["id"] = Rsp.Info.Id
		reMap["name"] = Rsp.Info.Name
		reMap["level"] = Rsp.Info.Level
		reMap["parent_category"] = Rsp.Info.ParentCategory
		reMap["is_tab"] = Rsp.Info.IsTab
		reMap["sub_categorys"] = Rsp.SubCategorys

		ctx.JSON(http.StatusOK, reMap)
	}
	return
}

func New(ctx *gin.Context) {
	var Categoryform forms.CategoryForm
	if err := ctx.ShouldBindJSON(&Categoryform); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	Rsp, err := global.GoodsSrvClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           Categoryform.Name,
		ParentCategory: Categoryform.ParentCategory,
		Level:          Categoryform.Level,
		IsTab:          *Categoryform.IsTab,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	//这里Grpc返回的时候，我只返回了Id,其他字段想要预览的话，就需要用Categoryform中的来赋值
	request := make(map[string]interface{})
	request["id"] = Rsp.Id
	request["name"] = Categoryform.Name
	request["parent"] = Categoryform.ParentCategory
	request["level"] = Categoryform.Level
	request["is_tab"] = *Categoryform.IsTab

	ctx.JSON(http.StatusOK, request)
}

func Delete(ctx *gin.Context) {
	categoryId := ctx.Param("id")
	id, err := strconv.ParseInt(categoryId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{
		Id: int32(id),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}

func Update(ctx *gin.Context) {
	updatecategoryform := forms.UpdateCategoryForm{}
	if err := ctx.ShouldBindJSON(&updatecategoryform); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	goodsId := ctx.Param("id")
	id, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	categoryinfoRequest := proto.CategoryInfoRequest{
		Id:   int32(id),
		Name: updatecategoryform.Name,
	}

	if updatecategoryform.IsTab != nil {
		categoryinfoRequest.IsTab = *updatecategoryform.IsTab
	}

	_, err = global.GoodsSrvClient.UpdateCategory(context.Background(), &categoryinfoRequest)
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.Status(http.StatusOK)
}
