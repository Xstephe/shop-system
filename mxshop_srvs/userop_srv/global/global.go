package global

import (
	"gorm.io/gorm"

	"mxshop_srvs/userop_srv/config"
)

var (
	DB           *gorm.DB
	ServerConfig config.ServerConfig
	NacosConfig  config.NacosConfig
)
