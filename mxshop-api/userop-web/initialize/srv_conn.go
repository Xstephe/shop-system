package initialize

import (
	"fmt"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"mxshop-api/userop-web/global"
	"mxshop-api/userop-web/proto"
)

func InitSrcConn() {
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【商品服务失败】")
	}
	//生成grpc的client并调用接口
	global.GoodsSrvClient = proto.NewGoodsClient(goodsConn)

	useropConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port, global.ServerConfig.UserOpSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatalf("[InitSrcConn] 连接 【用户操作服务失败】")
	}
	//生成grpc的client并调用接口
	MessageClient := proto.NewMessageClient(useropConn)
	global.MessageSrvClient = MessageClient

	AddressClient := proto.NewAddressClient(useropConn)
	global.AddressSrvClient = AddressClient

	UserFavClient := proto.NewUserFavClient(useropConn)
	global.UserFavSrvClient = UserFavClient

}
