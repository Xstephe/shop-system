package main

import (
	"fmt"
	"log"
	"mxshop_srvs/inventory_srv/model"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	dsn := "root:513513@tcp(192.168.182.130:3306)/mxshop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	//_ = db.AutoMigrate(&model.Inventory{}, &model.StockSellDetail{})

	//插入一条数据
	//orderDetail := model.StockSellDetail{
	//	OrderSn: "imooc bobby",
	//	Status:  1,
	//	Detail: []model.GoodsDetail{{
	//		Goods: 1,
	//		Num:   2,
	//	}},
	//}
	//db.Create(&orderDetail)

	var sellDetail model.StockSellDetail
	db.Where(model.StockSellDetail{OrderSn: "imooc bobby"}).First(&sellDetail)
	fmt.Println(sellDetail.Detail)
}
