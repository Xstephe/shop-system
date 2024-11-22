package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"
)

// 填充商品信息
func ModelToResponse(goods model.Goods) proto.GoodsInfoResponse {
	return proto.GoodsInfoResponse{
		Id:              goods.ID,
		CategoryId:      goods.CategoryID,
		Name:            goods.Name,
		GoodsSn:         goods.GoodsSn,
		ClickNum:        goods.ClickNum,
		SoldNum:         goods.SoldNum,
		FavNum:          goods.FavNum,
		MarketPrice:     goods.MarketPrice,
		ShopPrice:       goods.ShopPrice,
		GoodsBrief:      goods.GoodsBrief,
		ShipFree:        goods.ShipFree,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		OnSale:          goods.OnSale,
		DescImages:      goods.DescImages,
		Images:          goods.Images,
		GoodsDesc:       goods.Name, //这个定义的时候就少了个与之相对应的，设置成和名字一样
		Category: &proto.CategoryBriefInfoResponse{
			Id:   goods.Category.ID,
			Name: goods.Category.Name,
		},
		Brand: &proto.BrandInfoResponse{
			Id:   goods.Brands.ID,
			Name: goods.Brands.Name,
			Logo: goods.Brands.Logo,
		},
	}
}

// 获取商品列表
func (s *GoodsServer) GoodsList(ctx context.Context, req *proto.GoodsFilterRequest) (*proto.GoodsListResponse, error) {
	var goods []model.Goods
	var goodsListResponse proto.GoodsListResponse

	//定义一个临时变量
	//初始化筛选器
	q := elastic.NewBoolQuery()
	localDB := global.DB.Model(&model.Goods{})
	if req.KeyWords != "" {
		q = q.Must(elastic.NewMultiMatchQuery(req.KeyWords, "name", "goods_brief")) //在name和goods_brief查询
	}
	if req.IsNew {
		q = q.Filter(elastic.NewTermQuery("is_new", req.IsNew))
	}
	if req.IsHot {
		q = q.Filter(elastic.NewTermQuery("is_hot", req.IsHot))
	}
	if req.PriceMin > 0 {
		q = q.Filter(elastic.NewRangeQuery("shop_price").Gte(req.PriceMin))
	}
	if req.PriceMax > 0 {
		q = q.Filter(elastic.NewRangeQuery("shop_price").Lte(req.PriceMax))
	}
	if req.Brand > 0 {
		q = q.Filter(elastic.NewTermQuery("brands_id", req.Brand))
	}
	//通过category去查询商品
	// 子查询 嵌套 子查询
	// SELECT * FROM order WHERE category_id IN(SELECT id FROM category WHERE parent_category_id IN (SELECT id FROM category WHERE parent_category_id=1001))
	var subQuery string
	categoryIds := make([]interface{}, 0)
	if req.TopCategory > 0 {
		var category model.Category
		if result := global.DB.First(&category, req.TopCategory); result.RowsAffected == 0 {
			return nil, status.Errorf(codes.NotFound, "商品分类不存在")
		}
		if category.Level == 1 {
			subQuery = fmt.Sprintf("select id from category where parent_category_id in (select id from category WHERE parent_category_id=%d)", req.TopCategory)
		} else if category.Level == 2 {
			subQuery = fmt.Sprintf("select id from category WHERE parent_category_id=%d", req.TopCategory)
		} else if category.Level == 3 {
			subQuery = fmt.Sprintf("select id from category WHERE id=%d", req.TopCategory)
		}
		//这里需要将类目级别相关信息获取出相应的数据，用于在es查询相应商品信息
		type result struct {
			ID int32
		}
		var Results []result

		//将SubQuery语句获取到的Category映射到Result
		//获取对应分类的子分类id
		global.DB.Model(model.Category{}).Raw(subQuery).Scan(&Results)
		for _, p := range Results {
			categoryIds = append(categoryIds, p.ID)
		}
		//terms查询将目录级别分类条件放入q
		q = q.Filter(elastic.NewTermsQuery("category_id", categoryIds...))
	}

	//分页：From().Size()
	if req.Pages == 0 {
		req.Pages = 1
	}
	switch {
	case req.PagePerNums > 100:
		req.PagePerNums = 100
	case req.PagePerNums <= 0:
		req.PagePerNums = 10
	}

	//es根据query条件查询
	rsp, err := global.EsConfig.Search().Index(model.EsGoods{}.GetIndexName()).Query(q).From(int(req.Pages)).Size(int(req.PagePerNums)).Do(context.Background())
	if err != nil {
		return nil, err
	}

	//填充total
	goodsListResponse.Total = int32(rsp.Hits.TotalHits.Value)
	goodIDs := make([]int32, 0)
	//获取查询到的id
	for _, v := range rsp.Hits.Hits {
		goods := model.EsGoods{}
		json.Unmarshal(v.Source, &goods)
		goodIDs = append(goodIDs, goods.ID)
	}
	result := localDB.Preload("Category").Preload("Brands").Where("id IN ?", goodIDs).Find(&goods)
	if result.Error != nil {
		return nil, result.Error
	}
	for _, good := range goods {
		goodsInfoResponse := ModelToResponse(good)
		goodsListResponse.Data = append(goodsListResponse.Data, &goodsInfoResponse)
	}
	return &goodsListResponse, nil
}

// 批量查询商品的信息
func (s *GoodsServer) BatchGetGoods(ctx context.Context, req *proto.BatchGoodsIdInfo) (*proto.GoodsListResponse, error) {
	var goods []model.Goods
	var goodsListResponse proto.GoodsListResponse
	//调用where并不会真正执行sql 只是用来生成sql的 当调用find， first才会去执行sql，
	result := global.DB.Where(req.Id).Find(&goods)
	goodsListResponse.Total = int32(result.RowsAffected)
	for _, good := range goods {
		goodInfoResponse := ModelToResponse(good)
		goodsListResponse.Data = append(goodsListResponse.Data, &goodInfoResponse)
	}
	return &goodsListResponse, nil
}

// 创建商品
func (s *GoodsServer) CreateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*proto.GoodsInfoResponse, error) {
	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	//先检查redis中是否有这个token
	//防止同一个token的数据重复插入到数据库中，如果redis中没有这个token则放入redis
	//这里没有看到图片文件是如何上传， 在微服务中 普通的文件上传已经不再使用
	goods := model.Goods{
		Brands:          brand,
		BrandsID:        brand.ID,
		Category:        category,
		CategoryID:      category.ID,
		Name:            req.Name,
		GoodsSn:         req.GoodsSn,
		MarketPrice:     req.MarketPrice,
		ShopPrice:       req.ShopPrice,
		GoodsBrief:      req.GoodsBrief,
		ShipFree:        req.ShipFree,
		Images:          req.Images,
		DescImages:      req.DescImages,
		GoodsFrontImage: req.GoodsFrontImage,
		IsNew:           req.IsNew,
		IsHot:           req.IsHot,
		OnSale:          req.OnSale,
		Stocks:          req.Stocks,
	}
	//通过事务来保证mysql和es的一致性
	tx := global.DB.Begin()
	result := tx.Save(&goods)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error

	}
	tx.Commit()
	goodsInfoResponse := ModelToResponse(goods)
	return &goodsInfoResponse, nil
}

// 删除商品
func (s *GoodsServer) DeleteGoods(ctx context.Context, req *proto.DeleteGoodsInfo) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Goods{BaseModel: model.BaseModel{ID: req.Id}}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "要删除Id为%d的商品不存在", req.Id)
	}
	return &emptypb.Empty{}, nil
}

// 修改商品
func (s *GoodsServer) UpdateGoods(ctx context.Context, req *proto.CreateGoodsInfo) (*emptypb.Empty, error) {
	var good model.Goods

	if result := global.DB.First(&good, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品不存在")
	}

	var category model.Category
	if result := global.DB.First(&category, req.CategoryId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}

	var brand model.Brands
	if result := global.DB.First(&brand, req.BrandId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}

	good.Brands = brand
	good.BrandsID = brand.ID
	good.Category = category
	good.CategoryID = category.ID
	good.Name = req.Name
	good.GoodsSn = req.GoodsSn
	good.MarketPrice = req.MarketPrice
	good.ShopPrice = req.ShopPrice
	good.GoodsBrief = req.GoodsBrief
	good.ShipFree = req.ShipFree
	good.Images = req.Images
	good.DescImages = req.DescImages
	good.GoodsFrontImage = req.GoodsFrontImage
	good.IsNew = req.IsNew
	good.IsHot = req.IsHot
	good.OnSale = req.OnSale
	good.Stocks = req.Stocks

	tx := global.DB.Begin()
	result := tx.Save(&good)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

// 获取商品详情
func (s *GoodsServer) GetGoodsDetail(ctx context.Context, req *proto.GoodInfoRequest) (*proto.GoodsInfoResponse, error) {
	var good model.Goods
	if result := global.DB.Preload("Category").Preload("Brands").First(&good, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "Id为%d的商品不存在", req.Id)
	}
	goodInfoResponse := ModelToResponse(good)
	return &goodInfoResponse, nil
}
