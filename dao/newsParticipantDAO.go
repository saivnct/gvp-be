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

type NewsParticipantDAO struct {
	DAO
}

var singletonNewsParticipantDAO *NewsParticipantDAO
var onceNewsParticipantDAO sync.Once

func GetNewsParticipantDAO() *NewsParticipantDAO {
	onceNewsParticipantDAO.Do(func() {
		fmt.Println("Init NewsTagDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		newsParticipantDAO := NewsParticipantDAO{}
		newsParticipantDAO.Init(mongoCtx, &db.MongoDb)

		singletonNewsParticipantDAO = &newsParticipantDAO
	})
	return singletonNewsParticipantDAO
}

func (newsParticipantDAO *NewsParticipantDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "newsParticipants"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	newsParticipantDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (newsParticipantDAO *NewsParticipantDAO) Save(ctx context.Context, newsParticipant *model.NewsParticipant) (*model.NewsParticipant, error) {
	result := &model.NewsParticipant{}
	err := newsParticipantDAO.InsertOrUpdate(ctx, newsParticipant, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsParticipantDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (newsParticipantDAO *NewsParticipantDAO) FindByParticipant(ctx context.Context, participant string) (*model.NewsParticipant, error) {
	result := &model.NewsParticipant{}

	err := newsParticipantDAO.FindByPKey(ctx, participant, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsParticipantDAO - FindByParticipant): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (newsParticipantDAO *NewsParticipantDAO) UpdateByParticipant(ctx context.Context, participant string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.NewsParticipant, error) {
	result := &model.NewsParticipant{}

	err := newsParticipantDAO.UpdateByPKey(ctx, participant, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsParticipantDAO - UpdateByParticipant): failed executing UpdateOneByField -> %w", err)
	}

	return result, nil
}

func (newsParticipantDAO *NewsParticipantDAO) DeleteByParticipant(ctx context.Context, participant string) (*model.NewsParticipant, error) {
	result := &model.NewsParticipant{}
	err := newsParticipantDAO.DeleteByPKey(ctx, participant, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsParticipantDAO - DeleteByParticipant): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (newsParticipantDAO *NewsParticipantDAO) IncreaseListSearchCount(ctx context.Context, participants []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": participants,
		},
	}

	updateFields := primitive.M{}
	updateFields["searchCount"] = 1
	update := primitive.M{"$inc": updateFields}

	return newsParticipantDAO.UpdateMany(ctx, filter, update)
}

func (newsParticipantDAO *NewsParticipantDAO) IncreaseListNewsCount(ctx context.Context, participants []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": participants,
		},
	}

	updateFields := primitive.M{}
	updateFields["newsCount"] = 1
	update := primitive.M{"$inc": updateFields}

	return newsParticipantDAO.UpdateMany(ctx, filter, update)
}

func (newsParticipantDAO *NewsParticipantDAO) DecreaseListNewsCount(ctx context.Context, tags []string) (int64, error) {
	filter := primitive.M{
		PKEY_NAME: primitive.M{
			"$in": tags,
		},
	}

	updateFields := primitive.M{}
	updateFields["newsCount"] = -1
	update := primitive.M{"$inc": updateFields}

	return newsParticipantDAO.UpdateMany(ctx, filter, update)
}

func (newsParticipantDAO *NewsParticipantDAO) FetchListParticipants(ctx context.Context, page int64, pageSize int64, filter interface{}, sort interface{}) ([]*model.NewsParticipant, error) {
	results := []*model.NewsParticipant{}

	findOptions := options.Find()
	findOptions.SetSort(sort)
	if pageSize > 0 {
		findOptions.SetSkip(page * pageSize)
		findOptions.SetLimit(pageSize)
	}

	err := newsParticipantDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(NewsParticipantDAO - FetchListParticipants): failed executing FindAll -> %w", err)
	}

	return results, nil
}
