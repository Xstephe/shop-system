package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"mxshop-api/goods-web/global"
	"mxshop-api/goods-web/initialize"
	"mxshop-api/goods-web/utils"
	"mxshop-api/goods-web/utils/register/consul"
)

func main() {
	//初始化logger
	initialize.InitLogger()

	//初始化配置文件
	initialize.InitConfig()

	//初始化router
	Router := initialize.Routers()

	//初始化翻译器
	err := initialize.InitTrans("zh")
	if err != nil {
		panic(err)
	}
	//初始化srv的连接
	initialize.InitSrvConn()

	//初始化sentinel
	initialize.InitSentinel()

	//如果是本地开发环境固定端口号，线上环境启动获取端口号
	viper.AutomaticEnv()
	debug := viper.GetBool("MXSHOP_DEBUG")
	if !debug {
		port, err := utils.GetFreePost()
		if err == nil {
			global.ServerConfig.Port = port
		}

	}
	//consul配置
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, global.ServerConfig.Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("consul服务注册失败", err.Error())
	}

	zap.S().Debugf("启动服务器,端口:%d", global.ServerConfig.Port)

	go func() {
		err = Router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port))
		if err != nil {
			zap.S().Panic("启动失败", err.Error())
		}
	}()

	//接收终止信息
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.Deregister(serviceId); err != nil {
		zap.S().Info("注销失败")
	} else {
		zap.S().Info("注销成功")
	}
}
