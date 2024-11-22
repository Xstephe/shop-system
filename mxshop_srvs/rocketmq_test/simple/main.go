package main

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

func main() {
	//1.
	p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.182.130:9876"}))
	if err != nil {
		panic("生成producer失败")
	}
	//2.
	err = p.Start()
	if err != nil {
		panic("启动producer失败")
	}
	//3.同步
	res, err := p.SendSync(context.Background(), primitive.NewMessage("imooc", []byte(" this is imooc")))
	if err != nil {
		fmt.Printf("发送失败:%s\n", err)
	} else {
		fmt.Printf("发送成功:%s\n", res.String())
	}
	//4.
	err = p.Shutdown()
	if err != nil {
		panic("关闭producer失败")
	}
}
