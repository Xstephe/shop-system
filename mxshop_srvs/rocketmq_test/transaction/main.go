package main

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type OrderListener struct {
}

func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	fmt.Println("开始执行本地逻辑")
	time.Sleep(3 * time.Second)
	fmt.Println("执行本地逻辑失败")
	//本地执行逻辑无缘无故失败，代码宕机
	return primitive.UnknowState
}

// 回查
func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	fmt.Println("执行rocketmq的回查")
	time.Sleep(10 * time.Second)
	return primitive.CommitMessageState
}

func main() {
	p, err := rocketmq.NewTransactionProducer(&OrderListener{}, producer.WithNameServer([]string{"192.168.182.130:9876"}))
	if err != nil {
		fmt.Println("生成producer失败")
	}
	err = p.Start()
	if err != nil {
		fmt.Println("启动producer失败")
	}
	res, err := p.SendMessageInTransaction(context.Background(), primitive.NewMessage("TransTopic", []byte("this is transaction message3")))
	if err != nil {
		fmt.Println("发送失败")
	}
	fmt.Printf("发送成功:%s\n", res.String())

	time.Sleep(time.Hour)
	err = p.Shutdown()
	if err != nil {
		fmt.Println("关闭procedure失败")
	}
}
