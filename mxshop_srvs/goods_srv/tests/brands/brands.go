package main

import (
	"context"
	"fmt"

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

// 测试获取品牌列表
func TestBrandList() {
	rsp, err := goodsClient.BrandList(context.Background(), &proto.BrandFilterRequest{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)

	for _, UserInfoRsp := range rsp.Data {
		fmt.Println(UserInfoRsp.Name)
	}
}

// 测试创建品牌
func TestCreateBrand() {
	rsp, err := goodsClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: "杰士邦",
		Logo: "https://img30.jieshibang.com/popshop/jfs/t2428/114/2893255306/5394/6d2f141f/56f8a0feN621521bc.jpg",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

// 测试删除品牌
func TestDeleteBrand() {
	_, err := goodsClient.DeleteBrand(context.Background(), &proto.BrandRequest{
		Id: 1113,
	})
	if err != nil {
		panic(err)
	}
}

// 测试更新品牌
func TestUpdateBrand() {
	_, err := goodsClient.UpdateBrand(context.Background(), &proto.BrandRequest{
		Id:   1113,
		Name: "杰士邦",
		Logo: "https://img.jieshibang.com/popshop/jfs/t2428/114/2893255306/5394/6d2f141f/56f8a0feN621521bc.jpg",
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	Init()
	//TestBrandList()
	//TestCreateBrand()
	//TestDeleteBrand()
	//TestUpdateBrand()
	conn.Close()

}
