package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"
)

// 获取品牌列表
func (s *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	brandListResponse := &proto.BrandListResponse{}
	var brands []model.Brands
	//分页
	result := global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}
	var total int64
	global.DB.Model(&model.Brands{}).Count(&total)
	brandListResponse.Total = int32(total)

	var brandInfoResponse []*proto.BrandInfoResponse
	for _, brand := range brands {
		brandInfoResponse = append(brandInfoResponse, &proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		})
	}
	brandListResponse.Data = brandInfoResponse
	return brandListResponse, nil
}

// 新建品牌
func (s *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	//先查询品牌是否存在
	result := global.DB.Where("name=?", req.Name).First(&model.Brands{})
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.InvalidArgument, "要创建的品牌已存在")
	}
	var brand model.Brands
	brand.Name = req.Name
	brand.Logo = req.Logo
	global.DB.Save(&brand)
	brandInfoRsp := proto.BrandInfoResponse{
		Id:   brand.ID,
		Name: brand.Name,
		Logo: brand.Logo,
	}
	return &brandInfoRsp, nil
}

// 删除品牌
func (s *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	//先查询品牌是否存在
	result := global.DB.Delete(&model.Brands{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "要删除的品牌不存在")
	}
	return &emptypb.Empty{}, nil
}

// 更新品牌
func (s *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	//先查询品牌是否存在
	var brand model.Brands
	brand.ID = req.Id
	result := global.DB.First(&brand)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "要更新的品牌不存在")
	}
	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Logo != "" {
		brand.Logo = req.Logo
	}
	global.DB.Save(&brand)
	return &emptypb.Empty{}, nil
}
