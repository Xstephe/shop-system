package config

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ServerConfig struct {
	Tags       []string     `mapstructure:"tags" json:"tags"`
	Host       string       `mapstructure:"host" json:"host"`
	Name       string       `mapstructure:"name" json:"name"`
	MysqlInfo  MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	ConsulInfo ConsulConfig `mapstructure:"consul" json:"consul"`

	//商品微服务的配置
	GoodsSrvInfo GoodsSrvConfig `mapstructure:"goods_srv" json:"goods_srv"`
	//库存微服务的配置
	InventorySrvInfo GoodsSrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	DataID    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}
