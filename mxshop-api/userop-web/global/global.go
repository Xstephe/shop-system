package global

import (
	"mxshop-api/userop-web/config"
	"mxshop-api/userop-web/proto"

	ut "github.com/go-playground/universal-translator"
)

var (
	ServerConfig     *config.ServerConfig = &config.ServerConfig{}
	NacosConfig      *config.NacosConfig  = &config.NacosConfig{}
	Trans            ut.Translator
	GoodsSrvClient   proto.GoodsClient
	MessageSrvClient proto.MessageClient
	AddressSrvClient proto.AddressClient
	UserFavSrvClient proto.UserFavClient
)
