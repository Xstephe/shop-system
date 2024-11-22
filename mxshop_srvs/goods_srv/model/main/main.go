package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/olivere/elastic/v7"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"mxshop_srvs/goods_srv/model"
)

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

func main() {
	//dsn := "root:513513@tcp(192.168.182.130:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=true&loc=Local"
	////设置全局的logger，打印出sql语句
	//newLogger := logger.New(
	//	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	//	logger.Config{
	//		SlowThreshold:             time.Second, // Slow SQL threshold
	//		LogLevel:                  logger.Info, // Log level
	//		IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
	//		ParameterizedQueries:      false,       // Don't include params in the SQL log
	//		Colorful:                  true,        // Disable color
	//	},
	//)
	//db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
	//	Logger: newLogger,
	//	NamingStrategy: schema.NamingStrategy{
	//		SingularTable: true,
	//	},
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//_ = db.AutoMigrate(&model.Goods{}, &model.Brands{}, &model.Category{}, &model.Banner{}, &model.GoodsCategoryBrand{})
	Mysql2Es()
}

// 将mysql中的数据同步到es中
func Mysql2Es() {
	dsn := "root:513513@tcp(192.168.182.130:3306)/mxshop_goods_srv?charset=utf8mb4&parseTime=True&loc=Local"
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

	//_ = db.AutoMigrate(&model.Category{},
	//	&model.Brands{},
	//	&model.GoodsCategoryBrand{},
	//	&model.Banner{},
	//	&model.Goods{})

	//同步数据到es中
	host := "http://192.168.182.130:9200"
	logger := log.New(os.Stdout, "mxshop", log.LstdFlags)
	client, err := elastic.NewClient(elastic.SetURL(host), elastic.SetSniff(false), elastic.SetTraceLog(logger))
	if err != nil {
		panic(err)
	}
	var goodslist []model.Goods
	db.Find(&goodslist)
	for _, g := range goodslist {
		goods := model.EsGoods{
			ID:          g.ID,
			CategoryID:  g.CategoryID,
			BrandsID:    g.BrandsID,
			OnSale:      g.OnSale,
			ShipFree:    g.ShipFree,
			IsNew:       g.IsNew,
			IsHot:       g.IsHot,
			Name:        g.Name,
			ClickNum:    g.ClickNum,
			SoldNum:     g.SoldNum,
			FavNum:      g.FavNum,
			MarketPrice: g.MarketPrice,
			GoodsBrief:  g.GoodsBrief,
			ShopPrice:   g.ShopPrice,
		}

		_, err = client.Index().Index(goods.GetIndexName()).BodyJson(goods).Id(strconv.Itoa(int(goods.ID))).Do(context.Background())
		if err != nil {
			panic(err)
		}
	}
}
