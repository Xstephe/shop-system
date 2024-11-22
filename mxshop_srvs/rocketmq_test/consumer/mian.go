package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"time"
)

func main() {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.182.130:9876"}),
		consumer.WithGroupName("mxshop"))
	if err != nil {
		panic(err)
	}
	//订阅
	err = c.Subscribe("imooc", consumer.MessageSelector{}, func(ctx context.Context, msg ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msg {
			fmt.Printf("获取到值 %v\n", msg[i])
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Println("读取消息失败")
		panic(err)
	}
	_ = c.Start()
	//不能让主goroutine退出
	time.Sleep(time.Hour)
	_ = c.Shutdown()
}
