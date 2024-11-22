package main

import (
	"context"
	"fmt"
	"mxshop_srvs/order_srv/proto"

	"google.golang.org/grpc"
)

var ordClient proto.OrderClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	ordClient = proto.NewOrderClient(conn)
}

func TestCreateCartItem(userid, nums, goodsid int32) {
	rsp, err := ordClient.CreateCarItem(context.Background(), &proto.CartItemRequest{
		UserId:  userid,
		Nums:    nums,
		GoodsId: goodsid,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.Id)
}

func TestCartItemList(id int32) {
	rsp, err := ordClient.CartItemList(context.Background(), &proto.UserInfo{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range rsp.Data {
		fmt.Println(item.Id, item.GoodsId, item.Nums)
	}
}

func TestUpdateCartItem(id int32) {
	_, err := ordClient.UpdateCartItem(context.Background(), &proto.CartItemRequest{
		Id:      id,
		Checked: true,
	})
	if err != nil {
		panic(err)
	}
}

func TestCreateOrder() {
	_, err := ordClient.CreateOrder(context.Background(), &proto.OrderRequest{
		UserId:  2,
		Address: "陕西省",
		Name:    "韩信",
		Mobile:  "18736723156",
		Post:    "请尽快发货",
	})
	if err != nil {
		panic(err)
	}
}

func TestOrderDetail(id int32) {
	rsp, err := ordClient.OrderDetail(context.Background(), &proto.OrderRequest{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp.OrderInfo.OrderSn)
	for _, good := range rsp.Goods {
		fmt.Println(good.GoodsName)
	}
}

func TestOrderList() {
	rsp, err := ordClient.OrderList(context.Background(), &proto.OrderFilterRequest{
		UserId: 2,
	})
	if err != nil {
		panic(err)
	}
	for _, item := range rsp.Data {
		fmt.Println(item.OrderSn)
	}
}

func main() {
	Init()
	//TestCreateCartItem(2, 1, 422)
	//TestCartItemList(1)
	//TestUpdateCartItem(3)
	//TestCreateOrder()
	TestOrderDetail(1)
	//TestOrderList()
	//TestCreateOrder()
	conn.Close()

}
