package main

import (
	"flag"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/handler"
	"mxshop_srvs/inventory_srv/initialize"
	"mxshop_srvs/inventory_srv/proto"
	"mxshop_srvs/inventory_srv/utils"
	"mxshop_srvs/inventory_srv/utils/register/consul"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50053, "端口")

	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDb()
	initialize.InitRs()
	zap.S().Info(global.ServerConfig)

	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port: ", *Port)

	server := grpc.NewServer()
	proto.RegisterInventoryServer(server, &handler.InventoryServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		fmt.Println("failed to listen:", err.Error())
	}
	//注册服务健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//服务注册
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

	//监听库存归还topic
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.182.130:9876"}),
		consumer.WithGroupName("mxshop-inventory"))
	if err != nil {
		panic(err)
	}
	err = c.Subscribe("order_reback", consumer.MessageSelector{}, handler.AutoReback)
	if err != nil {
		fmt.Println("读取消息失败")
		panic(err)
	}
	_ = c.Start()

	//接收终止信息
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = c.Shutdown()
	if err = register_client.Deregister(serviceId); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")

}
