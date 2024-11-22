package api

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"mxshop-api/user-web/forms"
	"mxshop-api/user-web/global"
)

func GenerateSmsCode(witdh int) string {
	//生成width长度的短信验证码
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < witdh; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

// 发送短信,会把数据存储到redis中,
func SendSms(ctx *gin.Context) {
	//表单验证
	sendSmsForm := forms.SendSmsForm{}
	if err := ctx.ShouldBind(&sendSmsForm); err != nil {
		HandlerValidatorError(err, ctx)
		return
	}
	client, err := dysmsapi.NewClientWithAccessKey("cn-beijing", global.ServerConfig.AliSmsInfo.ApiKey, global.ServerConfig.AliSmsInfo.ApiSecret)
	if err != nil {
		panic(err)
	}
	smsCode := GenerateSmsCode(6)
	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" //https|http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-beijing"
	request.QueryParams["PhoneNumbers"] = sendSmsForm.Mobile            //手机号
	request.QueryParams["SignName"] = "然弟弟"                             //阿里云验证过的项目名
	request.QueryParams["TemplateCode"] = "SMS_305058098"               //阿里云的短信模板号
	request.QueryParams["TemplateParam"] = "{\"code\":" + smsCode + "}" //短信模板中的验证码内容
	response, err := client.ProcessCommonRequest(request)
	fmt.Print(client.DoAction(request, response))
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("response is %#v\n:", response)
	//将验证码保存起来  redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	rdb.Set(context.Background(), sendSmsForm.Mobile, smsCode, time.Duration(global.ServerConfig.AliSmsInfo.Expire)*time.Second)

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "发送成功",
	})

}
