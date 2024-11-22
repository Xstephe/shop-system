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

func TestCategoryBrandList() {
	rsp, err := goodsClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{
		Pages:       2,
		PagePerNums: 10,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	fmt.Println(rsp.Data)
}

func TestGetCategoryBrandList() {
	rsp, err := goodsClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id: 130366,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	fmt.Println(rsp.Data)
}

func main() {
	Init()
	TestCategoryBrandList()
	//TestGetCategoryBrandList()
	conn.Close()
}
