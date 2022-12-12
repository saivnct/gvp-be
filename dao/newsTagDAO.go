package dao

import (
	"context"
	"fmt"
	"gbb.go/gvp/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type NewsTagDAO struct {
	DAO
}

var singletonNewsTagDAO *NewsTagDAO
var onceNewsTagDAO sync.Once

func GetNewsTagDAO() *NewsTagDAO {
	onceNewsTagDAO.Do(func() {
		fmt.Println("Init NewsTagDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		newsTagDAO := NewsTagDAO{}
		newsTagDAO.Init(mongoCtx, &db.MongoDb)

		singletonNewsTagDAO = &newsTagDAO
	})
	return singletonNewsTagDAO
}

func (newsTagDAO *NewsTagDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "newsTags"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	newsTagDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (newsTagDAO *NewsTagDAO) Save(ctx context.Context, newsTag *model.NewsTag) (*model.NewsTag, error) {
	result := &model.NewsTag{}
	err := newsTagDAO.InsertOrUpdate(ctx, newsTag, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsTagDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (newsTagDAO *NewsTagDAO) FindByTag(ctx context.Context, tag string) (*model.NewsTag, error) {
	result := &model.NewsTag{}

	err := newsTagDAO.FindByPKey(ctx, tag, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsTagDAO - FindByTag): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (newsTagDAO *NewsTagDAO) UpdateByTag(ctx context.Context, tag string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.NewsTag, error) {
	result := &model.NewsTag{}

	err := newsTagDAO.UpdateByPKey(ctx, tag, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsTagDAO - UpdateByTag): failed executing UpdateOneByField -> %w", err)
	}

	return result, nil
}

func (newsTagDAO *NewsTagDAO) DeleteByTag(ctx context.Context, tag string) (*model.NewsTag, error) {
	result := &model.NewsTag{}
	err := newsTagDAO.DeleteByPKey(ctx, tag, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsTagDAO - DeleteByTag): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (newsTagDAO *NewsTagDAO) IncreaseListSearchCount(ctx context.Context, tags []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": tags,
		},
	}

	updateFields := primitive.M{}
	updateFields["searchCount"] = 1
	update := primitive.M{"$inc": updateFields}

	return newsTagDAO.UpdateMany(ctx, filter, update)
}

func (newsTagDAO *NewsTagDAO) IncreaseListNewsCount(ctx context.Context, tags []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": tags,
		},
	}

	updateFields := primitive.M{}
	updateFields["newsCount"] = 1
	update := primitive.M{"$inc": updateFields}

	return newsTagDAO.UpdateMany(ctx, filter, update)
}

func (newsTagDAO *NewsTagDAO) DecreaseListNewsCount(ctx context.Context, tags []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": tags,
		},
	}

	updateFields := primitive.M{}
	updateFields["newsCount"] = -1
	update := primitive.M{"$inc": updateFields}

	return newsTagDAO.UpdateMany(ctx, filter, update)
}

func (newsTagDAO *NewsTagDAO) FetchListTags(ctx context.Context, page int64, pageSize int64, filter interface{}, sort interface{}) ([]*model.NewsTag, error) {
	results := []*model.NewsTag{}

	findOptions := options.Find()
	findOptions.SetSort(sort)
	if pageSize > 0 {
		findOptions.SetSkip(page * pageSize)
		findOptions.SetLimit(pageSize)
	}

	err := newsTagDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(NewsTagDAO - FetchListTags): failed executing FindAll -> %w", err)
	}

	return results, nil
}
