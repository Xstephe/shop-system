package main

import (
	"flag"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"mxshop_srvs/order_srv/utils/otgrpc/otgrpc"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	uuid "github.com/satori/go.uuid"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/handler"
	"mxshop_srvs/order_srv/initialize"
	"mxshop_srvs/order_srv/proto"
	"mxshop_srvs/order_srv/utils"
	"mxshop_srvs/order_srv/utils/register/consul"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "ip地址")
	Port := flag.Int("port", 50051, "端口")

	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDb()
	initialize.InitSrvConn()
	zap.S().Info(global.ServerConfig)

	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port: ", *Port)

	//初始化trance
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1, // 50% 采样率
		},

		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "192.168.182.130:6831",
		},
		ServiceName: "mxshop",
	}
	tracer, closer, err := cfg.NewTracer(jaegercfg.Logger(jaegerlog.StdLogger))
	if err != nil {
		panic(err)
	}
	opentracing.SetGlobalTracer(tracer)
	server := grpc.NewServer(grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)))
	proto.RegisterOrderServer(server, &handler.OrderServer{})
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
			fmt.Println("failed to start grpc:", err.Error())
		}
	}()

	//监听订单超时topic
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.182.130:9876"}),
		consumer.WithGroupName("mxshop-order"))
	if err != nil {
		panic(err)
	}
	err = c.Subscribe("order_timeout", consumer.MessageSelector{}, handler.OrderTimeOut)
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
	_ = closer.Close()
	if err = register_client.Deregister(serviceId); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")
}
