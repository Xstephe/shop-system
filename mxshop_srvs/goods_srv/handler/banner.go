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

// 获取轮播图列表详情 不需要进行分页，因为数量很少
func (s *GoodsServer) BannerList(context.Context, *emptypb.Empty) (*proto.BannerListResponse, error) {
	var bannerListRsp proto.BannerListResponse
	var banners []model.Banner
	var bannerRsp []*proto.BannerResponse

	result := global.DB.Find(&banners)
	bannerListRsp.Total = int32(result.RowsAffected)
	for _, banner := range banners {
		bannerRsp = append(bannerRsp, &proto.BannerResponse{
			Id:    banner.ID,
			Index: banner.Index,
			Image: banner.Image,
			Url:   banner.Url,
		})
	}
	bannerListRsp.Data = bannerRsp
	return &bannerListRsp, nil
}

// 创建轮播图
func (s *GoodsServer) CreateBanner(ctx context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	//直接增加就行了，不用去查找存不存在，因为数量很少
	var banner model.Banner
	banner.Index = req.Index
	banner.Image = req.Image
	banner.Url = req.Url
	global.DB.Save(&banner)
	bannerRsp := &proto.BannerResponse{
		Id:    banner.ID,
		Index: banner.Index,
		Image: banner.Image,
		Url:   banner.Url,
	}
	return bannerRsp, nil
}

// 删除轮播图
func (s *GoodsServer) DeleteBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	result := global.DB.Delete(&model.Banner{}, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "要删除的轮播图不存在")
	}
	return &emptypb.Empty{}, nil
}

// 更新轮播图
func (s *GoodsServer) UpdateBanner(ctx context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	//先查找是否存在
	var banner model.Banner
	banner.ID = req.Id
	result := global.DB.First(&banner)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "要更新的轮播图不存在")
	}
	if req.Index != 0 {
		banner.Index = req.Index
	}
	if req.Url != "" {
		banner.Url = req.Url
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	global.DB.Save(&banner)
	return &emptypb.Empty{}, nil
}
