package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"mxshop_srvs/user_srv/proto"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic("dial err")
	}
	userClient = proto.NewUserClient(conn)
}

// 测试查询用户列表并校验密码
func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    2,
		PSize: 0,
	})
	if err != nil {
		panic(err)
	}
	for _, UserInfoRsp := range rsp.Data {
		fmt.Println(UserInfoRsp)
		checkRsp, err := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          "admin123",
			EncryptedPassword: UserInfoRsp.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkRsp.Success)
	}
	fmt.Println(rsp.Total)
}

// 测试创建用户
func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, err := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			NickName: fmt.Sprintf("bobby%d", i+9),
			Password: "admin123",
			Mobile:   fmt.Sprintf("1346291756%d", i),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp)
	}
}

// 通过手机号查询用户
func TestGetUserByMobile() {
	rsp, err := userClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: "18539857324",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

// 通过id查询用户
func TestGetUserById() {
	rsp, err := userClient.GetUserById(context.Background(), &proto.IdRequest{
		Id: 1,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(rsp)
}

// 更新用户信息
func TestUpdateUser() {
	// 需要传入的生日日期字符串
	dateStr := "2000-10-01"
	layout := "2006-01-02" // 日期格式

	// 解析日期字符串为 time.Time 类型
	t, _ := time.Parse(layout, dateStr)
	_, err := userClient.UpdateUser(context.Background(), &proto.UpdateUserInfo{
		Id:       11,
		NickName: "李蓬勃",
		Gender:   "male",
		BirthDay: uint64(t.Unix()),
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	Init()
	TestGetUserList()
	//TestCreateUser()
	//TestGetUserByMobile()
	//TestGetUserById()
	//TestUpdateUser()
	conn.Close()

}
