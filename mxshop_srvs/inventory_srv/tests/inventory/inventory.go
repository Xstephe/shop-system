package main

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"

	"mxshop_srvs/inventory_srv/proto"
)

var invClient proto.InventoryClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	invClient = proto.NewInventoryClient(conn)
}

func TestSetInv(goodsId, num int32) {
	_, err := invClient.SetInv(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodsId,
		Num:     num,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("set inv success")

}

func TestInvDetail(goodsId int32) {
	rsp, err := invClient.InvDetail(context.Background(), &proto.GoodsInvInfo{
		GoodsId: goodsId,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

func TestSell(wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := invClient.Sell(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodsInvInfo{
			{GoodsId: 421, Num: 1},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("库存扣减成功")
}

func TestReback() {
	_, err := invClient.Reback(context.Background(), &proto.SellInfo{
		GoodsInfo: []*proto.GoodsInvInfo{
			{GoodsId: 421, Num: 10},
			{GoodsId: 422, Num: 20},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("归还成功")

}

func main() {
	Init()
	//var i int32
	//for i = 421; i < 840; i++ {
	//	TestSetInv(i, 100)
	//}
	//TestSetInv(422, 110)
	//TestInvDetail(421)
	//TestSell()
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go TestSell(&wg)
	}
	wg.Wait()
	//TestReback()
	conn.Close()
}
