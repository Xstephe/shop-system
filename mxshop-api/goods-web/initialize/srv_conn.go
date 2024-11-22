package initialize

import (
	"fmt"
	"mxshop-api/goods-web/utils/otgrpc/otgrpc"

	"github.com/opentracing/opentracing-go"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/proto"
)

func InitSrvConn() {
	conn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.UserSrvConfig.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		zap.S().Errorw("[InitSrvConn]初始化连接 [商品服务] 失败")
		return
	}

	goodsSrvClient := proto.NewGoodsClient(conn)
	global.GoodsSrvClient = goodsSrvClient

}
