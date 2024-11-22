package global

import (
	"mxshop-api/order-web/config"
	"mxshop-api/order-web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	ServerConfig       *config.ServerConfig = &config.ServerConfig{}
	NacosConfig        *config.NacosConfig  = &config.NacosConfig{}
	Trans              ut.Translator
	GoodsSrvClient     proto.GoodsClient
	OrderSrvClient     proto.OrderClient
	InventorySrvClient proto.InventoryClient
)
