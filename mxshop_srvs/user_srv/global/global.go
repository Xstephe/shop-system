package global

import (
	"gorm.io/gorm"
	"mxshop_srvs/user_srv/config"
)

var (
	DB           *gorm.DB                                      //全局的操作数据库的变量
	ServerConfig *config.ServerConfig = &config.ServerConfig{} //全局的配置
	NacosConfig  *config.NacosConfig  = &config.NacosConfig{}
)
