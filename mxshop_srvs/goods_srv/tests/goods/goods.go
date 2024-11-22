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

// 测试获取商品列表
func TestGetGoodsList() {
	rsp, err := goodsClient.GoodsList(context.Background(), &proto.GoodsFilterRequest{
		TopCategory: 130361,
		//KeyWords:    "深海速冻",
		PriceMin: 90,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, good := range rsp.Data {
		fmt.Println(good.Name, good.ShopPrice, good.Brand.Name, good.Brand.Logo)
	}

}

// 测试批量获取商品
func TestBatchGetGoods() {
	rsp, err := goodsClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: []int32{421, 422, 423},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	for _, good := range rsp.Data {
		fmt.Println(good)
	}

}

// 测试获取商品详情
func TestGetGoodsDetail() {
	rsp, err := goodsClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: 421,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)

}

func main() {
	Init()
	//TestGetGoodsList()
	//TestGetGoodsDetail()
	TestBatchGetGoods()
	conn.Close()
}
