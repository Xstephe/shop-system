package handler

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"
)

// 获取所有商品分类列表
func (s *GoodsServer) GetAllCategorysList(context.Context, *emptypb.Empty) (*proto.CategoryListResponse, error) {
	/* 获取分类的时候构造好返回的json对象，供前端使用
		所以在CategoryListResponse结构中专门定义了JsonData用于返回给前端使用
		为什么在srv层来实现，因为srv层有gorm，而web层没有gorm并不与数据库交互
		如果在web层实现，没有gorm处理起来会比较复杂，所以建议放在srv层来实现
	[
		{
			"id":xxx,
			"name":"",
			"level":1,
			"is_tab":false,
			"parent":13xxx,
			"sub_category":[
				"id":xxx,
				"name":"",
				"level":1,
				"is_tab":false,
				"sub_category":[]
			]
		}
	]
	*/
	var categorys []model.Category
	//预加载出子目录
	result := global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)
	b, _ := json.Marshal(categorys)
	return &proto.CategoryListResponse{Total: int32(result.RowsAffected), JsonData: string(b)}, nil

}

// 获取子分类
func (s *GoodsServer) GetSubCategory(ctx context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var subcategoryListRsp proto.SubCategoryListResponse
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "查询的商品分类不存在")
	}
	//填充Info字段
	subcategoryListRsp.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}

	//填充SubCategorys字段
	var subCategorys []*proto.CategoryInfoResponse
	var subcategorys []model.Category
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Find(&subcategorys)
	for _, subcategoryInfo := range subcategorys {
		subCategorys = append(subCategorys, &proto.CategoryInfoResponse{
			Id:             subcategoryInfo.ID,
			Name:           subcategoryInfo.Name,
			ParentCategory: subcategoryInfo.ParentCategoryID,
			Level:          subcategoryInfo.Level,
			IsTab:          subcategoryInfo.IsTab,
		})
	}
	subcategoryListRsp.SubCategorys = subCategorys
	return &subcategoryListRsp, nil

}

// 创建品牌分类
func (s *GoodsServer) CreateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	var category model.Category
	category.Name = req.Name
	category.Level = req.Level
	category.IsTab = req.IsTab
	if req.Level != 1 {
		//查询父类目是否存在
		var checkcategory model.Category
		if result := global.DB.Where(&model.Category{ParentCategoryID: req.ParentCategory}).First(&checkcategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "新建品牌不存在父类目")
		}
		category.ParentCategoryID = req.ParentCategory
	}
	global.DB.Create(&category)
	categoryInfo := &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}
	return categoryInfo, nil

}

// 删除品牌分类
func (s *GoodsServer) DeleteCategory(ctx context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {

	var category model.Category
	if result := global.DB.Delete(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	return &emptypb.Empty{}, nil
}

// 修改品牌分类
func (s *GoodsServer) UpdateCategory(ctx context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	//先查询商品分类存不存在
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory != 0 {
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab {
		category.IsTab = req.IsTab
	}
	global.DB.Save(&category)
	return &emptypb.Empty{}, nil
}
