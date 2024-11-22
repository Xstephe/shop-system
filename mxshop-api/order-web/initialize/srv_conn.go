package initialize

import (
	"fmt"
	"mxshop-api/order-web/utils/otgrpc/otgrpc"

	"github.com/opentracing/opentracing-go"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"mxshop-api/order-web/global"
	"mxshop-api/order-web/proto"
)

func InitSrcConn() {
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【商品服务失败】")
	}
	//生成grpc的client并调用接口
	global.GoodsSrvClient = proto.NewGoodsClient(goodsConn)

	orderConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.OrderSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【订单服务失败】")
	}
	//生成grpc的client并调用接口
	global.OrderSrvClient = proto.NewOrderClient(orderConn)

	InvConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.InventorySrv.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【库存服务失败】")
	}
	//生成grpc的client并调用接口
	global.InventorySrvClient = proto.NewInventoryClient(InvConn)
}
