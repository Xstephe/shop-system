package main

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc"

	"mxshop_srvs/goods_srv/proto"
)

var goodsClient proto.GoodsClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic("dial err")
	}
	goodsClient = proto.NewGoodsClient(conn)
}

// 测试获取轮播图列表
func TestBannerList() {
	rsp, err := goodsClient.BannerList(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)

	for _, BannerInfoRsp := range rsp.Data {
		fmt.Println(BannerInfoRsp)
	}
}

// 测试创建轮播图
func TestCreateBanner() {
	rsp, err := goodsClient.CreateBanner(context.Background(), &proto.BannerRequest{
		Index: 3,
		Image: "http://shop.cchjwtsd.com/media/banner/banner3_u5SU5y2.jpg",
		Url:   "423",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

// 测试删除轮播图
func TestDeleteBrand() {
	_, err := goodsClient.DeleteBanner(context.Background(), &proto.BannerRequest{
		Id: 6,
	})
	if err != nil {
		panic(err)
	}
}

// 测试更新轮播图
func TestUpdateBanner() {
	_, err := goodsClient.UpdateBanner(context.Background(), &proto.BannerRequest{
		Id:    6,
		Image: "http://shop.cchjwtsd.com/media/banner/banner3.jpg",
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	Init()
	//TestBannerList()
	//TestCreateBanner()
	//TestDeleteBrand()
	TestUpdateBanner()
	conn.Close()

}
