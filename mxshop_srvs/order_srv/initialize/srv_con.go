package initialize

import (
	"fmt"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/proto"
)

func InitSrvConn() {
	//初始化商品服务连接
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【商品服务失败】")
	}
	//生成grpc的client并调用接口
	goodsSrcClient := proto.NewGoodsClient(goodsConn)
	global.GoodsSrvClient = goodsSrcClient

	//初始化库存服务连接
	invConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【库存服务失败】")
	}
	//生成grpc的client并调用接口
	invSrcClient := proto.NewInventoryClient(invConn)
	global.InventorySrvClient = invSrcClient
}
