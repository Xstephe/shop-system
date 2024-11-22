package config

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type JwtConfig struct {
	SigningKey string `mapstructure:"key" json:"key"`
}

type JaegerConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
	Name string `mapstructure:"name" json:"name"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type AlipayConfig struct {
	AppId        string `mapstructure:"app_id" json:"app_id"`
	PrivateKey   string `mapstructure:"private_key" json:"private_key"`
	AliPublicKey string `mapstructure:"ali_public_key" json:"ali_public_key"`
	NotifyUrl    string `mapstructure:"notify_url" json:"notify_url"`
	ReturnUrl    string `mapstructure:"return_url" json:"return_url"`
}

type ServerConfig struct {
	Name         string         `mapstructure:"name" json:"name"`
	Host         string         `mapstructure:"host" json:"host"`
	Port         int            `mapstructure:"port" json:"port"`
	Tags         []string       `mapstructure:"tags" json:"tags"`
	GoodsSrvInfo GoodsSrvConfig `mapstructure:"goods_srv" json:"goods_srv"`
	OrderSrvInfo GoodsSrvConfig `mapstructure:"order_srv" json:"order_srv"`
	InventorySrv GoodsSrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
	JWTInfo      JwtConfig      `mapstructure:"jwt" json:"jwt"`
	ConsulInfo   ConsulConfig   `mapstructure:"consul" json:"consul"`
	AliPayInfo   AlipayConfig   `mapstructure:"alipay" json:"alipay"`
	JaegerInfo   JaegerConfig   `mapstructure:"jaeger" json:"jaeger"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	DataID    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}
