package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

type Registry struct {
	Host string
	Port int
}
type RegistryClient interface {
	Register(address string, port int, name string, tags []string, id string) error
	Deregister(serviceId string) error
}

func NewRegistryClient(host string, port int) RegistryClient {
	return &Registry{host, port}
}

func (r *Registry) Register(address string, port int, name string, tags []string, id string) error {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		zap.S().Error("创建consul实例失败")
		return err
	}
	//生成对应的检查对象
	check := &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d", address, port),
		Timeout:                        "10s", // 尝试增加超时时间
		Interval:                       "10s", // 尝试增加检查间隔
		DeregisterCriticalServiceAfter: "5s",  // 延长注销时间
	}

	//生成注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = name
	registration.Address = address
	registration.Port = port
	registration.Tags = tags
	registration.ID = id
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		zap.S().Error("consul服务注册失败")
		return err
	}
	zap.S().Info("consul服务注册成功")
	return nil
}

func (r *Registry) Deregister(serviceId string) error {
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", r.Host, r.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		zap.S().Error("创建consul实例失败")
		return err
	}
	err = client.Agent().ServiceDeregister(serviceId)
	return err
}
