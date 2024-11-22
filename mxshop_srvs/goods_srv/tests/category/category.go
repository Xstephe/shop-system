package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"

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

// 测试获取品牌分类列表详情
func TestGetAllCategorysList() {
	rsp, err := goodsClient.GetAllCategorysList(context.Background(), &empty.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Total)
	fmt.Println(rsp.JsonData)
}

// 测试品牌下的子分类
func TestGetSubCategory() {
	rsp, err := goodsClient.GetSubCategory(context.Background(), &proto.CategoryListRequest{
		Id: 130364,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Info)
	fmt.Println(rsp.SubCategorys)
}

// 测试新建品牌分类
func TestCreateCategory() {
	rsp, err := goodsClient.CreateCategory(context.Background(), &proto.CategoryInfoRequest{
		Name:           "红苹果",
		ParentCategory: 45641241,
		Level:          3,
		IsTab:          false,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

// 测试更新品牌分类
func TestUpdateCategory() {
	_, err := goodsClient.UpdateCategory(context.Background(), &proto.CategoryInfoRequest{
		Id:    238010,
		Name:  "小苹果",
		IsTab: true,
	})
	if err != nil {
		panic(err)
	}
}

// 测试删除品牌分类
func TestUpdateBrand() {
	_, err := goodsClient.DeleteCategory(context.Background(), &proto.DeleteCategoryRequest{
		Id: 238010,
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	Init()
	//TestGetAllCategorysList()
	//TestGetSubCategory()
	TestCreateCategory()
	//TestUpdateBrand()
	//TestUpdateCategory()
	conn.Close()

}
