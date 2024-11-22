package brands

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mxshop-api/goods-web/api"
	"mxshop-api/goods-web/forms"
	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/proto"
)

// BrandList 获取品牌列表
func List(ctx *gin.Context) {
	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)

	rsp, err := global.GoodsSrvClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages:       int32(pnInt),
		PagePerNums: int32(pSizeInt),
	})

	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]interface{}, 0)
	reMap := make(map[string]interface{})
	reMap["total"] = rsp.Total
	//[pnInt : pnInt*pSizeInt+pSizeInt]
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["name"] = value.Name
		reMap["logo"] = value.Logo
		result = append(result, reMap)
	}

	reMap["data"] = result

	ctx.JSON(http.StatusOK, reMap)
}

// NewBrand 新建品牌
func New(ctx *gin.Context) {
	BrandForm := forms.BrandForm{}
	if err := ctx.ShouldBindJSON(&BrandForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	Rsp, err := global.GoodsSrvClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: BrandForm.Name,
		Logo: BrandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	Response := make(map[string]interface{})
	Response["id"] = Rsp.Id
	Response["name"] = BrandForm.Name
	Response["loge"] = BrandForm.Logo

	ctx.JSON(http.StatusOK, Response)
}

// DeleteBrand 删除品牌
func Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

// UpdateBrand 更新品牌
func Update(ctx *gin.Context) {
	UpdateBrandForm := forms.BrandForm{}
	if err := ctx.ShouldBindJSON(&UpdateBrandForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	brandId := ctx.Param("id")
	id, err := strconv.ParseInt(brandId, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateBrand(context.Background(), &proto.BrandRequest{
		Id:   int32(id),
		Name: UpdateBrandForm.Name,
		Logo: UpdateBrandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}

// GetCategoryBrand 获取分类-品牌详情
func GetCategoryBrand(ctx *gin.Context) {
	cbid := ctx.Param("id")
	id, err := strconv.ParseInt(cbid, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	Rsp, err := global.GoodsSrvClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: int32(id),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, Rsp.Data)
}

// GetCategoryBrandList 获取分类-品牌列表，获取分类下的所有品牌
func GetCategoryBrandList(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	rsp, err := global.GoodsSrvClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: int32(i),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["name"] = value.Name
		reMap["logo"] = value.Logo

		result = append(result, reMap)
	}

	ctx.JSON(http.StatusOK, result)
}

// CategoryBrandList 获取分类-品牌表关系
func CategoryBrandList(ctx *gin.Context) {

	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)

	rsp, err := global.GoodsSrvClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{
		Pages:       int32(pnInt),
		PagePerNums: int32(pSizeInt),
	})

	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	reMap := map[string]interface{}{
		"total": rsp.Total,
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["category"] = map[string]interface{}{
			"id":   value.Category.Id,
			"name": value.Category.Name,
		}
		reMap["brand"] = map[string]interface{}{
			"id":   value.Brand.Id,
			"name": value.Brand.Name,
			"logo": value.Brand.Logo,
		}

		result = append(result, reMap)
	}

	reMap["data"] = result
	ctx.JSON(http.StatusOK, reMap)
}

// NewCategoryBrand 新建分类-品牌关系
func NewCategoryBrand(ctx *gin.Context) {
	categoryBrandForm := forms.CategoryBrandForm{}
	if err := ctx.ShouldBindJSON(&categoryBrandForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	rsp, err := global.GoodsSrvClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: int32(categoryBrandForm.CategoryId),
		BrandId:    int32(categoryBrandForm.BrandId),
	})

	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	response := make(map[string]interface{})
	response["id"] = rsp.Id

	ctx.JSON(http.StatusOK, response)
}

// UpdateCategoryBrand 更新分类-品牌关系
func UpdateCategoryBrand(ctx *gin.Context) {
	categoryBrandForm := forms.CategoryBrandForm{}
	if err := ctx.ShouldBindJSON(&categoryBrandForm); err != nil {
		api.HandlerValidatorError(err, ctx)
		return
	}

	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id:         int32(i),
		CategoryId: int32(categoryBrandForm.CategoryId),
		BrandId:    int32(categoryBrandForm.BrandId),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "修改成功",
	})
}

// DeleteCategoryBrand 删除分类-品牌关系
func DeleteCategoryBrand(ctx *gin.Context) {
	id := ctx.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteCategoryBrand(context.Background(), &proto.CategoryBrandRequest{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}
