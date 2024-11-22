package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/model"
	"mxshop_srvs/inventory_srv/proto"
)

type InventoryServer struct {
	proto.UnimplementedInventoryServer
}

// 设置库存
func (*InventoryServer) SetInv(ctx context.Context, req *proto.GoodsInvInfo) (*emptypb.Empty, error) {
	var inv model.Inventory
	global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv)
	inv.Goods = req.GoodsId
	inv.Stocks = req.Num
	global.DB.Save(&inv)
	return &emptypb.Empty{}, nil
}

// 库存详情
func (*InventoryServer) InvDetail(ctx context.Context, req *proto.GoodsInvInfo) (*proto.GoodsInvInfo, error) {
	var inv model.Inventory
	if result := global.DB.Where(&model.Inventory{Goods: req.GoodsId}).First(&inv); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
	}
	return &proto.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

// 扣减库存
func (*InventoryServer) Sell(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//需要使用到事务的特性
	tx := global.DB.Begin() //开启事务

	sellDetail := model.StockSellDetail{
		OrderSn: req.OrderSn,
		Status:  1,
	}
	var details []model.GoodsDetail
	//使用redis分布式锁
	mutex := global.Rs.NewMutex(fmt.Sprintf("421"))
	for _, goodInfo := range req.GoodsInfo {
		details = append(details, model.GoodsDetail{
			Goods: goodInfo.GoodsId,
			Num:   goodInfo.Num,
		})
		var inv model.Inventory
		//使用悲观锁
		//if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//	tx.Rollback() //事务回滚
		//	return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		//}

		//使用乐观锁
		//for {
		//	if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
		//		tx.Rollback() //事务回滚
		//		return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		//	}
		//	//判断库存是否充足
		//	if inv.Stocks < goodInfo.Num {
		//		tx.Rollback() //事务回滚
		//		return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		//	}
		//	//扣减
		//	inv.Stocks -= goodInfo.Num
		//
		//	//乐观锁更新语句
		//	//零值，int字段默认是0，对于gorm来说，gorm会默认忽略  需要加上Select字段，强制指明哪个字段更新
		//	if result := tx.Model(&model.Inventory{}).Select("stocks", "version").Where("goods = ? and version = ?", goodInfo.GoodsId, inv.Version).Updates(&model.Inventory{Stocks: inv.Stocks, Version: inv.Version + 1}); result.RowsAffected == 0 {
		//		zap.S().Info("更新库存失败")
		//	} else {
		//		break
		//	}
		//
		//}

		if err := mutex.Lock(); err != nil {
			return nil, status.Errorf(codes.Internal, "获取redis锁失败")
		}

		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //事务回滚
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		//判断库存是否充足
		if inv.Stocks < goodInfo.Num {
			tx.Rollback() //事务回滚
			return nil, status.Errorf(codes.ResourceExhausted, "库存不足")
		}
		//扣减
		inv.Stocks -= goodInfo.Num

		tx.Save(&inv)

	}
	sellDetail.Detail = details
	//写selldetail表
	if result := tx.Create(&sellDetail); result.RowsAffected == 0 {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "保存库存扣减历史失败")
	}
	tx.Commit() //提交事务
	if ok, err := mutex.Unlock(); !ok || err != nil {
		return nil, status.Errorf(codes.Internal, "释放redis锁失败")
	}
	return &emptypb.Empty{}, nil
}
func (*InventoryServer) Reback(ctx context.Context, req *proto.SellInfo) (*emptypb.Empty, error) {
	//库存归还 1：订单超时 2：订单创建失败 3：手动归还
	//需要使用到事务的特性
	tx := global.DB.Begin()
	for _, goodInfo := range req.GoodsInfo {
		var inv model.Inventory
		if result := global.DB.Where(&model.Inventory{Goods: goodInfo.GoodsId}).First(&inv); result.RowsAffected == 0 {
			tx.Rollback() //事务回滚
			return nil, status.Errorf(codes.NotFound, "库存信息不存在")
		}
		//要加分布式锁
		inv.Stocks += goodInfo.Num
		tx.Save(&inv)
	}

	tx.Commit() //提交事务
	return &emptypb.Empty{}, nil
}

func AutoReback(ctx context.Context, msg ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	type OrderInfo struct {
		OrderSn string
	}
	for i := range msg {
		//既然是归还库存，那么我应该具体地知道每件商品应该归还多少
		//这个接口应该确保幂等性，不能反复归还
		//需要新建一张表，来记录订单的扣减细节

		var orderInfo OrderInfo
		err := json.Unmarshal(msg[i].Body, &orderInfo)
		if err != nil {
			zap.S().Errorf("解析json失败：%s", msg[i].Body)
			return consumer.ConsumeSuccess, nil
		}

		//库存扣减记录
		var stockSellDetail model.StockSellDetail
		//开始事务
		tx := global.DB.Begin()

		//获取需要归还的库存
		if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: orderInfo.OrderSn, Status: 1}).First(&stockSellDetail); result.RowsAffected == 0 {
			return consumer.ConsumeSuccess, nil
		}
		//如果查询到，逐个归还
		for _, orderGoods := range stockSellDetail.Detail {
			//先查询inventory, update语句update xxx set stocks=stocks+2 当多个并发进入mysql会自动锁住，更安全
			if result := tx.Model(&model.Inventory{}).Where(&model.Inventory{Goods: orderGoods.Goods}).Update("stocks", gorm.Expr("stocks+?", orderGoods.Num)); result.RowsAffected == 0 {
				tx.Rollback()
				//过一段时间重新消费ConsumeRetryLater
				return consumer.ConsumeRetryLater, nil
			}

			//更新状态
			if result := tx.Model(&model.StockSellDetail{}).Where(&model.StockSellDetail{OrderSn: stockSellDetail.OrderSn}).Update("status", 2); result.RowsAffected == 0 {
				tx.Rollback()
				return consumer.ConsumeRetryLater, nil
			}
		}
		tx.Commit()

		return consumer.ConsumeSuccess, nil
	}
	return consumer.ConsumeSuccess, nil
}
