package dao

import (
	"context"
	"fmt"
	"gbb.go/gvp/model"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type UserTmpDAO struct {
	DAO
}

var singletonUserTmpDAO *UserTmpDAO
var onceUserTmpDAO sync.Once

func GetUserTmpDAO() *UserTmpDAO {
	onceUserTmpDAO.Do(func() {
		fmt.Println("Init UserTmpDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		userTmpDAO := UserTmpDAO{}
		userTmpDAO.Init(mongoCtx, &db.MongoDb)

		singletonUserTmpDAO = &userTmpDAO
	})
	return singletonUserTmpDAO
}

func (userTmpDAO *UserTmpDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "userTmps"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	userTmpDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (userTmpDAO *UserTmpDAO) Save(ctx context.Context, userTmp *model.UserTmp) (*model.UserTmp, error) {
	result := &model.UserTmp{}
	err := userTmpDAO.InsertOrUpdate(ctx, userTmp, result)
	if err != nil {
		return nil, fmt.Errorf("(UserTmpDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (userTmpDAO *UserTmpDAO) FindByEmail(ctx context.Context, email string) (*model.UserTmp, error) {
	result := &model.UserTmp{}

	err := userTmpDAO.FindByPKey(ctx, email, result)
	if err != nil {
		return nil, fmt.Errorf("(UserTmpDAO - FindByEmail): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (userTmpDAO *UserTmpDAO) DeleteByEmail(ctx context.Context, email string) (*model.UserTmp, error) {
	result := &model.UserTmp{}
	err := userTmpDAO.DeleteByPKey(ctx, email, result)
	if err != nil {
		return nil, fmt.Errorf("(UserTmpDAO - DeleteByEmail): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (userTmpDAO *UserTmpDAO) UpdateByEmail(ctx context.Context, email string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.UserTmp, error) {
	result := &model.UserTmp{}

	err := userTmpDAO.UpdateByPKey(ctx, email, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - UpdateByEmail): failed executing UpdateOneByField -> %w", err)
	}
	return result, nil
}
