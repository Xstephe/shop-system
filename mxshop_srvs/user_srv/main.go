package main

import (
	"flag"
	"fmt"
	"mxshop_srvs/user_srv/utils/register/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"mxshop_srvs/user_srv/global"
	"mxshop_srvs/user_srv/handler"
	"mxshop_srvs/user_srv/initialize"
	"mxshop_srvs/user_srv/proto"
	"mxshop_srvs/user_srv/utils"
)

func main() {

	//手动设置IP和端口号
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50052, "端口号")
	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()

	flag.Parse()
	zap.S().Info("ip:", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePost()
	}
	zap.S().Info("port:", *Port)

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//consul服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败", err.Error())
	}
	// zap.S()可以获取一个全局的sugar，可以让我们自己设置一个全局的logger
	zap.S().Debugf("启动服务器，端口：%d", *Port)
	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to serve: " + err.Error())
		}
	}()

	//优雅退出consul
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.Deregister(serviceId); err != nil {
		zap.S().Info("注销失败")
	} else {
		zap.S().Info("注销成功")
	}
}
