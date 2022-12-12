package dao

import (
	"context"
	"fmt"
	"gbb.go/gvp/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type CategoryDAO struct {
	DAO
}

var singletonCategoryDAO *CategoryDAO
var onceCategoryDAO sync.Once

func GetCategoryDAO() *CategoryDAO {
	onceCategoryDAO.Do(func() {
		fmt.Println("Init CategoryDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		categoryDAO := CategoryDAO{}
		categoryDAO.Init(mongoCtx, &db.MongoDb)

		singletonCategoryDAO = &categoryDAO
	})
	return singletonCategoryDAO
}

func (categoryDAO *CategoryDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "categories"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	categoryDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (categoryDAO *CategoryDAO) Save(ctx context.Context, category *model.Category) (*model.Category, error) {
	result := &model.Category{}
	err := categoryDAO.InsertOrUpdate(ctx, category, result)
	if err != nil {
		return nil, fmt.Errorf("(CategoryDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (categoryDAO *CategoryDAO) FindByCatId(ctx context.Context, catId string) (*model.Category, error) {
	result := &model.Category{}

	err := categoryDAO.FindByPKey(ctx, catId, result)
	if err != nil {
		return nil, fmt.Errorf("(CategoryDAO - FindByCatId): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (categoryDAO *CategoryDAO) UpdateByCatId(ctx context.Context, catId string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.Category, error) {
	result := &model.Category{}

	err := categoryDAO.UpdateByPKey(ctx, catId, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(CategoryDAO - UpdateByCatId): failed executing UpdateOneByField -> %w", err)
	}

	return result, nil
}

func (categoryDAO *CategoryDAO) DeleteByCatId(ctx context.Context, catId string) (*model.Category, error) {
	result := &model.Category{}
	err := categoryDAO.DeleteByPKey(ctx, catId, result)
	if err != nil {
		return nil, fmt.Errorf("(CategoryDAO - DeleteByCatId): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (categoryDAO *CategoryDAO) GetAll(ctx context.Context) ([]*model.Category, error) {
	results := []*model.Category{}

	filter := bson.M{}

	findOptions := options.Find()
	findOptions.SetSort(primitive.M{"createdAt": 1})

	err := categoryDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(CategoryDAO - GetAll): failed executing SearchByField -> %w", err)
	}

	return results, nil

}
