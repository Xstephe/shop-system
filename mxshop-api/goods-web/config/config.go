package config

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"key" json:"key"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type JaegerConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
	Name string `mapstructure:"name" json:"name"`
}

type ServerConfig struct {
	Name          string         `mapstructure:"name" json:"name"`
	Host          string         `mapstructure:"host" json:"host"`
	Port          int            `mapstructure:"port" json:"port"`
	Tags          []string       `mapstructure:"tags" json:"tags"`
	UserSrvConfig GoodsSrvConfig `mapstructure:"goods_srv" json:"goods_srv"`
	JWTInfo       JWTConfig      `mapstructure:"jwt" json:"jwt"`
	ConsulInfo    ConsulConfig   `mapstructure:"consul" json:"consul"`
	JaegerInfo    JaegerConfig   `mapstructure:"jaeger" json:"jaeger"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	DataID    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}
