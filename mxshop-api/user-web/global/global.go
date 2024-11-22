package global

import (
	ut "github.com/go-playground/universal-translator"

	"mxshop-api/user-web/config"
	"mxshop-api/user-web/proto"
)

var (
	// 设置中文错误翻译器
	Trans ut.Translator

	ServerConfig *config.ServerConfig = &config.ServerConfig{} //Server配置实例

	NacosConfig *config.NacosConfig = &config.NacosConfig{} //Nacos配置实例

	UserSrvClient proto.UserClient
)
