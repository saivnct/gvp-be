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

type NewsCommentDAO struct {
	DAO
}

var singletonNewsCommentDAO *NewsCommentDAO
var onceNewsCommentDAO sync.Once

func GetNewsCommentDAO() *NewsCommentDAO {
	onceNewsCommentDAO.Do(func() {
		fmt.Println("Init NewsCommentDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		newsCommentDAO := NewsCommentDAO{}
		newsCommentDAO.Init(mongoCtx, &db.MongoDb)

		singletonNewsCommentDAO = &newsCommentDAO
	})
	return singletonNewsCommentDAO
}

func (newsCommentDAO *NewsCommentDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "newsComments"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	newsCommentDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (newsCommentDAO *NewsCommentDAO) Save(ctx context.Context, newsComment *model.NewsComment) (*model.NewsComment, error) {
	result := &model.NewsComment{}
	err := newsCommentDAO.InsertOrUpdate(ctx, newsComment, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsCommentDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (newsCommentDAO *NewsCommentDAO) FindByCommentId(ctx context.Context, commentId string) (*model.NewsComment, error) {
	result := &model.NewsComment{}

	err := newsCommentDAO.FindByPKey(ctx, commentId, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsCommentDAO - FindByCommentId): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (newsCommentDAO *NewsCommentDAO) UpdateByCommentId(ctx context.Context, commentId string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.NewsComment, error) {
	result := &model.NewsComment{}

	err := newsCommentDAO.UpdateByPKey(ctx, commentId, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsCommentDAO - UpdateByCommentId): failed executing UpdateOneByField -> %w", err)
	}

	return result, nil
}

func (newsCommentDAO *NewsCommentDAO) DeleteByCommentId(ctx context.Context, commentId string) (*model.NewsComment, error) {
	result := &model.NewsComment{}
	err := newsCommentDAO.DeleteByPKey(ctx, commentId, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsCommentDAO - DeleteByCommentId): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (newsCommentDAO *NewsCommentDAO) FetchListNewsCommentsFilter(newsId string, parentCommentId string) interface{} {
	/*
		{
			$and: [
				"newsId": newsId,
				$expr:{
					$eq: [{"$arrayElemAt": ["$commentAncestors", -1]}, parentCommentId]
				}
			]
		}
	*/

	filter := primitive.M{}

	filterAnds := []interface{}{}
	if len(newsId) > 0 {
		filterAnds = append(filterAnds, primitive.M{
			"newsId": newsId,
		})
	}

	if len(parentCommentId) > 0 {
		//pre-last element of $commentAncestors is parentCommentId => get only 1 level from parentCommentId
		filterAnds = append(filterAnds, primitive.M{
			"$expr": primitive.M{
				"$eq": primitive.A{
					primitive.M{
						"$arrayElemAt": primitive.A{"$commentAncestors", -2},
					},
					parentCommentId,
				},
			},
		})
	} else {
		filterAnds = append(filterAnds, primitive.M{
			"commentAncestors": primitive.M{
				"$size": 1,
			},
		})
	}

	if len(filterAnds) > 0 {
		filter["$and"] = filterAnds
	}

	return filter
}

func (newsCommentDAO *NewsCommentDAO) CountListNewsComments(ctx context.Context, newsId string, parentCommentId string) (int64, error) {
	filter := newsCommentDAO.FetchListNewsCommentsFilter(newsId, parentCommentId)

	totalItem, err := newsCommentDAO.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("(NewsDAO - CountListNewsComments): failed executing CountDocuments -> %w", err)
	}

	return totalItem, nil
}

func (newsCommentDAO *NewsCommentDAO) FetchListNewsComments(ctx context.Context, page int64, pageSize int64, newsId string, parentCommentId string) ([]*model.NewsComment, error) {
	results := []*model.NewsComment{}

	filter := newsCommentDAO.FetchListNewsCommentsFilter(newsId, parentCommentId)

	findOptions := options.Find()
	findOptions.SetSort(primitive.M{"createdAt": -1})
	if pageSize > 0 {
		findOptions.SetSkip(page * pageSize)
		findOptions.SetLimit(pageSize)
	}

	err := newsCommentDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - FetchListNewsComments): failed executing FindAll -> %w", err)
	}

	return results, nil
}
