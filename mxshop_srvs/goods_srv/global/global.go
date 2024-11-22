package global

import (
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
	"mxshop_srvs/goods_srv/config"
)

var (
	DB           *gorm.DB                                      //全局的操作数据库的变量
	ServerConfig *config.ServerConfig = &config.ServerConfig{} //全局的配置
	NacosConfig  *config.NacosConfig  = &config.NacosConfig{}
	EsConfig     *elastic.Client
)
