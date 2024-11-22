package model

import (
	"database/sql/driver"
	"encoding/json"
)

//type Stock struct {
//	BaseModel
//	Name string
//	Address string
//}

type Inventory struct {
	BaseModel
	Goods   int32 `gorm:"type:int;index"`
	Stocks  int32 `gorm:"type:int"`
	Version int32 `gorm:"type:int"` //分布式锁的乐观锁
}

// GoodsDetail 商品详细
type GoodsDetail struct {
	Goods int32 //商品id
	Num   int32 //商品数量
}

// GoodsDetailList gorm自定义模型
type GoodsDetailList []GoodsDetail

func (g GoodsDetailList) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GoodsDetailList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}

// StockSellDetail 库存扣减细节 记录扣减库存历史记录
type StockSellDetail struct {
	OrderSn string          `gorm:"type:varchar(200);index:idx_order_sn,unique"` //订单号uuid，平台自己生成的订单号
	Status  int32           `gorm:"type:varchar(200)"`                           //1 表示库存已扣减  2 表示库存已归还
	Detail  GoodsDetailList `gorm:"type:varchar(200)"`
}

func (StockSellDetail) TableName() string {
	return "stockselldetail"
}
