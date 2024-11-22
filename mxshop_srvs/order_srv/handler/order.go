package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"math/rand"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"

	"go.uber.org/zap"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/model"
	"mxshop_srvs/order_srv/proto"
)

func GenerateOrderSn(userID int32) string {
	//订单号的生成规则
	/*
	 年月日时分秒+用户id+2位随机数
	*/

	now := time.Now()
	rand.Seed(now.UnixNano())
	orderSn := fmt.Sprintf("%d%d%d%d%d%d%d%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Nanosecond(),
		userID, rand.Intn(90)+10)
	return orderSn
}

func (*OrderServer) CartItemList(ctx context.Context, req *proto.UserInfo) (*proto.CartItemListResponse, error) {
	//获取用户的购物车列表
	var shopCarts []model.ShoppingCart
	var rsp proto.CartItemListResponse

	if result := global.DB.Where(&model.ShoppingCart{User: req.Id}).Find(&shopCarts); result.Error != nil {
		return nil, result.Error
	} else {
		rsp.Total = int32(result.RowsAffected)
	}
	for _, shopCart := range shopCarts {
		rsp.Data = append(rsp.Data, &proto.ShopCartInfoResponse{
			Id:      shopCart.ID,
			UserId:  shopCart.User,
			GoodsId: shopCart.Goods,
			Nums:    shopCart.Nums,
			Checked: shopCart.Checked,
		})
	}
	return &rsp, nil
}
func (*OrderServer) CreateCarItem(ctx context.Context, req *proto.CartItemRequest) (*proto.ShopCartInfoResponse, error) {
	//添加商品到购物车中 1. 购物车没有这件商品- 新建一个记录 ，2. 这个商品之前添加到了购物车，合并
	var shopCart model.ShoppingCart
	if result := global.DB.Where(&model.ShoppingCart{Goods: req.GoodsId, User: req.UserId}).First(&shopCart); result.RowsAffected == 0 {
		//没有这条记录
		shopCart.Goods = req.GoodsId
		shopCart.User = req.UserId
		shopCart.Nums = req.Nums
		shopCart.Checked = false
	} else {
		//记录已经存在，合并购物车记录
		shopCart.Nums += req.Nums
	}
	global.DB.Save(&shopCart)
	return &proto.ShopCartInfoResponse{Id: shopCart.ID}, nil

}
func (*OrderServer) UpdateCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	//更新购物车信息 更新数量和选中状态
	var shopCart model.ShoppingCart
	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).First(&shopCart); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	shopCart.Checked = req.Checked
	if req.Nums > 0 {
		shopCart.Nums = req.Nums
	}
	global.DB.Save(&shopCart)
	return &emptypb.Empty{}, nil
}
func (*OrderServer) DeleteCartItem(ctx context.Context, req *proto.CartItemRequest) (*emptypb.Empty, error) {
	//删除购物车信息
	if result := global.DB.Where("goods=? and user=?", req.GoodsId, req.UserId).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "购物车记录不存在")
	}
	return &emptypb.Empty{}, nil
}

type OrderListener struct {
	Code        codes.Code      //返回状态码
	Detail      string          //返回状态描述
	ID          int32           //订单id
	OrderAmount float32         //订单总金额
	Ctx         context.Context //上下文数据
}

func (o *OrderListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)
	parentSpan := opentracing.SpanFromContext(o.Ctx)

	var shopCarts []model.ShoppingCart
	var goodsId []int32
	goodsNumsMap := make(map[int32]int32)
	shopCartSpan := opentracing.GlobalTracer().StartSpan("select_shopcart", opentracing.ChildOf(parentSpan.Context()))
	if result := global.DB.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Find(&shopCarts); result.RowsAffected == 0 {
		o.Code = codes.InvalidArgument
		o.Detail = "没有选中结算的商品"
		return primitive.RollbackMessageState
	}
	shopCartSpan.Finish()

	for _, shopCart := range shopCarts {
		goodsId = append(goodsId, shopCart.Goods)
		goodsNumsMap[shopCart.Goods] = shopCart.Nums
	}

	//跨商品服务调用
	queryGoodsSpan := opentracing.GlobalTracer().StartSpan("query_goods", opentracing.ChildOf(parentSpan.Context()))
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{Id: goodsId})
	if err != nil {
		o.Code = codes.Internal
		o.Detail = "批量查询商品信息失败"
		return primitive.RollbackMessageState
	}
	queryGoodsSpan.Finish()

	var orderAmount float32
	var orderGoods []*model.OrderGoods
	var goodsInvInfo []*proto.GoodsInvInfo
	for _, good := range goods.Data {
		orderAmount += good.ShopPrice * float32(goodsNumsMap[good.Id])
		orderGoods = append(orderGoods, &model.OrderGoods{
			Goods:      good.Id,
			GoodsName:  good.Name,
			GoodsImage: good.GoodsFrontImage,
			GoodsPrice: good.ShopPrice,
			Nums:       goodsNumsMap[good.Id],
		})

		goodsInvInfo = append(goodsInvInfo, &proto.GoodsInvInfo{
			GoodsId: good.Id,
			Num:     goodsNumsMap[good.Id],
		})
	}

	//跨库存扣减，库存服务
	queryInvSpan := opentracing.GlobalTracer().StartSpan("query_inv", opentracing.ChildOf(parentSpan.Context()))
	if _, err := global.InventorySrvClient.Sell(context.Background(), &proto.SellInfo{
		OrderSn:   orderInfo.OrderSn,
		GoodsInfo: goodsInvInfo,
	}); err != nil {
		o.Code = codes.ResourceExhausted
		o.Detail = "扣减库存失败"
		return primitive.RollbackMessageState
	}
	queryInvSpan.Finish()

	tx := global.DB.Begin()
	//生成订单表
	orderInfo.OrderMount = orderAmount
	saveOrderSpan := opentracing.GlobalTracer().StartSpan("save_order", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.Save(&orderInfo); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "创建订单失败"
		return primitive.CommitMessageState
	}
	saveOrderSpan.Finish()

	o.OrderAmount = orderAmount
	o.ID = orderInfo.ID
	for _, orderGood := range orderGoods {
		orderGood.Order = orderInfo.ID
	}

	//批量插入orderGoods
	saveOrderGoodsSpan := opentracing.GlobalTracer().StartSpan("save_order_goods", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.CreateInBatches(orderGoods, 100); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "批量插入商品订单失败"
		return primitive.CommitMessageState
	}
	saveOrderGoodsSpan.Finish()

	//删除
	deleteShopCartSpan := opentracing.GlobalTracer().StartSpan("delete_shopcart", opentracing.ChildOf(parentSpan.Context()))
	if result := tx.Where(&model.ShoppingCart{User: orderInfo.User, Checked: true}).Delete(&model.ShoppingCart{}); result.RowsAffected == 0 {
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "删除购物车记录失败"
		return primitive.CommitMessageState
	}
	deleteShopCartSpan.Finish()

	//发送延时消息
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
	message := primitive.NewMessage("order_timeout", msg.Body)
	message.WithDelayTimeLevel(5)
	res, err := p.SendSync(context.Background(), message)
	if err != nil {
		zap.S().Errorf("发送延时消息失败：%v\n", err)
		tx.Rollback()
		o.Code = codes.Internal
		o.Detail = "发送延时消息失败"
		return primitive.CommitMessageState
	} else {
		fmt.Printf("发送成功:%s\n", res.String())
	}

	//提交事务
	tx.Commit()
	o.Code = codes.OK
	return primitive.RollbackMessageState
}

// 回查
func (o *OrderListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var orderInfo model.OrderInfo
	_ = json.Unmarshal(msg.Body, &orderInfo)

	//怎么检查之前的逻辑是否完成
	if result := global.DB.Where(model.OrderInfo{OrderSn: orderInfo.OrderSn}).First(&orderInfo); result.RowsAffected == 0 {
		return primitive.CommitMessageState //你并不能说明这里库存就是扣减了
	}
	return primitive.RollbackMessageState
}

func (*OrderServer) CreateOrder(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoResponse, error) {
	/*
		新建订单
			1. 从购物车中获取选中的商品
			2. 商品的金额自己查询 -访问商品服务（跨微服务）
			3. 库存的扣减 -访问库存服务（跨微服务）
			4. 订单的基本信息表  - 订单的商品信息表
			5. 从购物车中删除已购买的记录
	*/
	orderListener := OrderListener{Ctx: ctx}
	p, err := rocketmq.NewTransactionProducer(&orderListener,
		producer.WithNameServer([]string{"192.168.182.130:9876"}))
	if err != nil {
		zap.S().Errorf("生成producer失败: %s", err.Error())
		return nil, err
	}
	err = p.Start()
	if err != nil {
		zap.S().Errorf("启动producer失败: %s", err.Error())
		return nil, err
	}

	order := model.OrderInfo{
		OrderSn:      GenerateOrderSn(req.UserId),
		Address:      req.Address,
		SignerName:   req.Name,
		SingerMobile: req.Mobile,
		Post:         req.Post,
		User:         req.UserId,
	}
	jsonString, _ := json.Marshal(order)

	_, err = p.SendMessageInTransaction(context.Background(),
		primitive.NewMessage("order_reback", jsonString))
	if err != nil {
		zap.S().Error("发送失败")
		return nil, status.Error(codes.Internal, "发送消息失败")
	}
	if orderListener.Code != codes.OK {
		return nil, status.Error(codes.Internal, orderListener.Detail)
	}

	return &proto.OrderInfoResponse{Id: orderListener.ID, OrderSn: order.OrderSn, Total: orderListener.OrderAmount}, nil
}
func (*OrderServer) OrderList(ctx context.Context, req *proto.OrderFilterRequest) (*proto.OrderListResponse, error) {
	//订单列表
	var orders []model.OrderInfo
	var rsp proto.OrderListResponse

	var total int64
	global.DB.Where(&model.OrderInfo{User: req.UserId}).Find(&model.OrderInfo{}).Count(&total)
	rsp.Total = int32(total)

	//分页
	global.DB.Scopes(Paginate(int(req.Pages), int(req.Pages))).Where(&model.OrderInfo{User: req.UserId}).Find(&orders)
	for _, order := range orders {
		rsp.Data = append(rsp.Data, &proto.OrderInfoResponse{
			Id:      order.ID,
			UserId:  order.User,
			OrderSn: order.OrderSn,
			PayType: order.PayType,
			Status:  order.Status,
			Post:    order.Post,
			Total:   order.OrderMount,
			Address: order.Address,
			Name:    order.SignerName,
			Mobile:  order.SingerMobile,
			AddTime: order.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &rsp, nil
}
func (*OrderServer) OrderDetail(ctx context.Context, req *proto.OrderRequest) (*proto.OrderInfoDetailResponse, error) {
	//订单详情
	var order model.OrderInfo
	var rsp proto.OrderInfoDetailResponse
	var ordergoods []model.OrderGoods
	if result := global.DB.Where(&model.OrderInfo{BaseModel: model.BaseModel{ID: req.Id}, User: req.UserId}).First(&order); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	orderInfo := proto.OrderInfoResponse{}
	orderInfo.Id = order.ID
	orderInfo.UserId = order.User
	orderInfo.OrderSn = order.OrderSn
	orderInfo.PayType = order.PayType
	orderInfo.Status = order.Status
	orderInfo.Post = order.Post
	orderInfo.Total = order.OrderMount
	orderInfo.Address = order.Address
	orderInfo.Name = order.SignerName
	orderInfo.Mobile = order.SingerMobile
	orderInfo.AddTime = order.CreatedAt.Format("2006-01-02 15:04:05")

	rsp.OrderInfo = &orderInfo

	if result := global.DB.Where(&model.OrderGoods{Order: order.ID}).Find(&ordergoods); result.Error != nil {
		return nil, result.Error
	}
	for _, ordergood := range ordergoods {
		rsp.Goods = append(rsp.Goods, &proto.OrderItemResponse{
			GoodsId:    ordergood.Goods,
			GoodsName:  ordergood.GoodsName,
			GoodsImage: ordergood.GoodsImage,
			GoodsPrice: ordergood.GoodsPrice,
			Nums:       ordergood.Nums,
			OrderId:    ordergood.Order,
			Id:         order.ID,
		})
	}
	return &rsp, nil
}
func (*OrderServer) UpdateOrderStatus(ctx context.Context, req *proto.OrderStatus) (*emptypb.Empty, error) {
	//更新订单的状态
	if result := global.DB.Model(&model.OrderInfo{}).Where("order_sn = ?", req.OrderSn).Update("status", req.Status); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "订单不存在")
	}
	return &emptypb.Empty{}, nil
}

func OrderTimeOut(ctx context.Context, msg ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	for i := range msg {
		var OrderInfo model.OrderInfo
		_ = json.Unmarshal(msg[i].Body, &OrderInfo)

		//如果支付成功，什么都不做，超时支付则归还库存
		var order model.OrderInfo
		if result := global.DB.Model(&model.OrderInfo{}).Where(&model.OrderInfo{OrderSn: OrderInfo.OrderSn}).First(&order); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}

		fmt.Println("订单状态", order.Status)

		if order.Status != "TRADE_SUCCESS" {
			//归还库存，模仿order发送消息到order-reback中

			//事务
			tx := global.DB.Begin()

			//更改订单状态支付成功
			order.Status = "TRADE_CLOSED"
			tx.Save(&order)

			p, err := rocketmq.NewProducer(producer.WithNameServer([]string{"192.168.182.130:9876"}))
			if err != nil {
				panic("生成producer失败")
			}

			err = p.Start()
			if err != nil {
				panic("启动producer失败")
			}

			mq := primitive.NewMessage("order-reback", msg[i].Body)
			_, err = p.SendSync(context.Background(), mq)
			if err != nil {
				tx.Rollback()
				zap.S().Errorf("发送失败%s", err)
				return consumer.ConsumeRetryLater, nil
			}

			tx.Commit()
			return consumer.ConsumeSuccess, nil
		}
	}
	return consumer.ConsumeSuccess, nil
}
