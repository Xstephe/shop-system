package initialize

import (
	"encoding/json"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"mxshop_srvs/user_srv/global"
)

func GetEnvInfo(env string) bool {
	viper.AutomaticEnv()
	return viper.GetBool(env)
}

// 从配置文件中读取对应的配置
func InitConfig() {
	debug := GetEnvInfo("MXSHOP_DEBUG")
	configFilePrefix := "config"
	configFile := fmt.Sprintf("user_srv/%s-pro.yaml", configFilePrefix)
	if debug {
		configFile = fmt.Sprintf("user_srv/%s-debug.yaml", configFilePrefix)
	}
	v := viper.New()
	//文件的路径
	v.SetConfigFile(configFile)
	err := v.ReadInConfig()
	if err != nil {
		zap.S().Errorf("读取配置文件失败")
		return
	}
	_ = v.Unmarshal(global.NacosConfig)
	zap.S().Infof("配置信息%v", *global.NacosConfig)

	////viper的动态监控
	//v.WatchConfig()
	//v.OnConfigChange(func(e fsnotify.Event) {
	//	zap.S().Infof("配置文件产生变化%s", e.Name)
	//	_ = v.ReadInConfig()
	//	_ = v.Unmarshal(global.ServerConfig)
	//	zap.S().Infof("配置信息%v", *global.ServerConfig)
	//})

	//从Nacos中读取配置
	sc := []constant.ServerConfig{
		{
			IpAddr: global.NacosConfig.Host,
			Port:   global.NacosConfig.Port,
		},
	}
	cc := constant.ClientConfig{
		NamespaceId:         global.NacosConfig.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// 创建动态配置客户端
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		panic(err)
	}

	// 尝试获取配置
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: global.NacosConfig.DataID,
		Group:  global.NacosConfig.Group,
	})
	if err != nil {
		panic(err)
	}

	//fmt.Println(content)

	//想要将一个json字符串转换成struct，需要去设置这个struct的tag
	err = json.Unmarshal([]byte(content), &global.ServerConfig)
	if err != nil {
		zap.S().Errorf("读取nacos配置失败:%s", err.Error())
		return
	}
}
