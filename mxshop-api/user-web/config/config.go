package config

type UserSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"key" json:"key"`
}

type AliSmsConfig struct {
	ApiKey    string `mapstructure:"key" json:"key"`
	ApiSecret string `mapstructure:"secret" json:"secret"`
	Expire    int    `mapstructure:"expire" json:"expire"`
}

type RedisConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type ServerConfig struct {
	Name          string        `mapstructure:"name" json:"name"`
	Port          int           `mapstructure:"port" json:"port"`
	Host          string        `mapstructure:"host" json:"host"`
	Tags          []string      `mapstructure:"tags" json:"tags"`
	UserSrvConfig UserSrvConfig `mapstructure:"user_srv" json:"user_srv"`
	JWTInfo       JWTConfig     `mapstructure:"jwt" json:"jwt"`
	AliSmsInfo    AliSmsConfig  `mapstructure:"sms" json:"sms"`
	RedisInfo     RedisConfig   `mapstructure:"redis" json:"redis"`
	ConsulInfo    ConsulConfig  `mapstructure:"consul" json:"consul"`
}

type NacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	DataID    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}
