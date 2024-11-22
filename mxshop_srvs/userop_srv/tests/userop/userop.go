package main

import (
	"context"
	"mxshop_srvs/userop_srv/proto"

	"google.golang.org/grpc"
)

var UserFavClient proto.UserFavClient
var messageClient proto.MessageClient
var addressClient proto.AddressClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	UserFavClient = proto.NewUserFavClient(conn)
	messageClient = proto.NewMessageClient(conn)
	addressClient = proto.NewAddressClient(conn)
}

func TestAddressList() {
	_, err := addressClient.GetAddressList(context.Background(), &proto.GetAddressRequest{
		UserId: 1,
	})
	if err != nil {
		panic(err)
	}
}

func TestMessageList() {
	_, err := messageClient.GetMessageList(context.Background(), &proto.MessageRequest{
		UserId: 1,
	})
	if err != nil {
		panic(err)
	}
}

func TestUserFavList() {
	_, err := UserFavClient.GetFavList(context.Background(), &proto.UserFavRequest{
		UserId: 1,
	})
	if err != nil {
		panic(err)
	}
}

func main() {
	Init()
	TestAddressList()
	TestMessageList()
	TestUserFavList()
	conn.Close()
}
