package initialize

import (
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"go.uber.org/zap"
)

func InitSentinel() {
	//初始化sentinel
	err := sentinel.InitDefault()
	if err != nil {
		zap.S().Fatalf("初始化sentinel%v", err)
	}
	//配置限流规则
	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "goods-list",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              3,
			StatIntervalInMs:       5000,
		},
	})
	if err != nil {
		zap.S().Fatalf("加载规则失败: %+v", err)
		return
	}

}
