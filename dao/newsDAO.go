package dao

import (
	"context"
	"fmt"
	"gbb.go/gvp/model"
	"gbb.go/gvp/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

type NewsDAO struct {
	DAO
}

var singletonNewsDAO *NewsDAO
var onceNewsDAO sync.Once

func GetNewsDAO() *NewsDAO {
	onceNewsDAO.Do(func() {
		fmt.Println("Init NewsDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		newsDAO := NewsDAO{}
		newsDAO.Init(mongoCtx, &db.MongoDb)

		singletonNewsDAO = &newsDAO
	})
	return singletonNewsDAO
}

func (newsDAO *NewsDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "news"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	newsDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (newsDAO *NewsDAO) Save(ctx context.Context, news *model.News) (*model.News, error) {
	result := &model.News{}
	err := newsDAO.InsertOrUpdate(ctx, news, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) FindByNewsId(ctx context.Context, newsId string) (*model.News, error) {
	result := &model.News{}

	err := newsDAO.FindByPKey(ctx, newsId, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - FindByNewsId): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) UpdateByNewsId(ctx context.Context, newsId string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.News, error) {
	result := &model.News{}

	err := newsDAO.UpdateByPKey(ctx, newsId, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - UpdateByNewsId): failed executing UpdateOneByField -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) DeleteByNewsId(ctx context.Context, newsId string) (*model.News, error) {
	result := &model.News{}
	err := newsDAO.DeleteByPKey(ctx, newsId, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - DeleteByNewsId): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (newsDAO *NewsDAO) CountByCatId(ctx context.Context, catId string) (int64, error) {
	filter := bson.M{"categories": catId}

	result, err := newsDAO.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("(NewsDAO - CountByCatId): failed executing CountDocuments -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) FetchListNewsFilter(catIds []string, tags []string, participants []string, searchPhrase string, author string) interface{} {
	/*
		{
			$and: [
				"author" : author,
				"categories": {
					$in: catIds
				},
				"tags": {
					"$in": tags,
				},
				"participants": {
					"$in": participants,
				},
				$or : [
					"title": {
						$regex: ".*searchPhrase.*",
						$options: "i"
					},

					"participants": {
						$regex: ".*searchPhrase.*",
						$options: "i"
					},
				],
			]
		}
	*/

	filter := primitive.M{}

	filterAnds := []interface{}{}
	if len(author) > 0 {
		filterAnds = append(filterAnds, primitive.M{
			"author": author,
		})
	}

	if len(catIds) > 0 {
		filterAnds = append(filterAnds, primitive.M{
			"categories": primitive.M{
				"$in": catIds,
			},
		})
	}

	if len(tags) > 0 {
		filterAnds = append(filterAnds, primitive.M{
			"tags": primitive.M{
				"$in": tags,
			},
		})
	}

	if len(participants) > 0 {
		filterAnds = append(filterAnds, primitive.M{
			"participants": primitive.M{
				"$in": participants,
			},
		})
	}

	if len(searchPhrase) > 0 {
		pattern := fmt.Sprintf(".*%v.*", searchPhrase)
		search := bson.M{
			"$regex":   pattern,
			"$options": "i",
		}

		orArr := []interface{}{
			primitive.M{"title": search},
			primitive.M{"participants": search},
		}

		filterAnds = append(filterAnds, primitive.M{
			"$or": orArr,
		})
	}

	if len(filterAnds) > 0 {
		filter["$and"] = filterAnds
	}

	return filter
}

func (newsDAO *NewsDAO) CountListNews(ctx context.Context, catIds []string, tags []string, participants []string, searchPhrase string, author string) (int64, error) {
	filter := newsDAO.FetchListNewsFilter(catIds, tags, participants, searchPhrase, author)

	totalItem, err := newsDAO.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("(NewsDAO - CountListNews): failed executing CountDocuments -> %w", err)
	}

	return totalItem, nil
}

func (newsDAO *NewsDAO) FetchListNews(ctx context.Context, page int64, pageSize int64, catIds []string, tags []string, participants []string, searchPhrase string, author string, sort interface{}) ([]*model.News, error) {
	results := []*model.News{}

	filter := newsDAO.FetchListNewsFilter(catIds, tags, participants, searchPhrase, author)

	findOptions := options.Find()
	findOptions.SetSort(sort)
	if pageSize > 0 {
		findOptions.SetSkip(page * pageSize)
		findOptions.SetLimit(pageSize)
	}

	err := newsDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - FetchListNews): failed executing FindAll -> %w", err)
	}

	return results, nil
}

func (newsDAO *NewsDAO) AppendPreviewImage(ctx context.Context, newsId string, fileId string) (*model.News, error) {
	result := &model.News{}

	updateFields := primitive.M{}
	updateFields["previewImages"] = fileId
	update := primitive.M{"$addToSet": updateFields}

	err := newsDAO.UpdateByPKey(ctx, newsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - AppendPreviewImage): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) RemovePreviewImage(ctx context.Context, newsId string, fileId string) (*model.News, error) {
	result := &model.News{}

	//$pull, which "removes all instances of a value from an existing array"

	updateFields := primitive.M{}
	updateFields["previewImages"] = fileId
	update := primitive.M{"$pull": updateFields}

	err := newsDAO.UpdateByPKey(ctx, newsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - RemovePreviewImage): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) AppendMedia(ctx context.Context, newsId string, fileId string) (*model.News, error) {
	result := &model.News{}

	updateFields := primitive.M{}
	updateFields["medias"] = fileId
	update := primitive.M{"$addToSet": updateFields}

	err := newsDAO.UpdateByPKey(ctx, newsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - AppendMedia): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) RemoveMedia(ctx context.Context, newsId string, fileId string) (*model.News, error) {
	result := &model.News{}

	//$pull, which "removes all instances of a value from an existing array"

	updateFields := primitive.M{}
	updateFields["medias"] = fileId
	update := primitive.M{"$pull": updateFields}

	err := newsDAO.UpdateByPKey(ctx, newsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - RemoveMedia): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) IncreaseViews(ctx context.Context, news *model.News) (*model.News, error) {
	result := &model.News{}

	update := primitive.M{}

	updateSetFields := primitive.M{}

	updateIncFields := primitive.M{}
	updateIncFields["views"] = 1

	currentWeek := utils.UTCNowBeginningOfWeek()
	if currentWeek == news.CurrentViewsWeek {
		updateIncFields["weekViews"] = 1
	} else {
		updateSetFields["weekViews"] = 1
		updateSetFields["currentViewsWeek"] = currentWeek
	}

	currentMonth := utils.UTCNowBeginningOfMonth()
	if currentMonth == news.CurrentViewsMonth {
		updateIncFields["monthViews"] = 1
	} else {
		updateSetFields["monthViews"] = 1
		updateSetFields["currentViewsMonth"] = currentMonth
	}

	update["$inc"] = updateIncFields
	update["$set"] = updateSetFields

	err := newsDAO.UpdateByPKey(ctx, news.NewsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - IncreaseViews): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) IncreaseLikes(ctx context.Context, news *model.News, likeBy *model.User) (*model.News, error) {
	result := &model.News{}

	update := primitive.M{}

	updateSetFields := primitive.M{}

	updateIncFields := primitive.M{}
	updateIncFields["likes"] = 1

	currentWeek := utils.UTCNowBeginningOfWeek()
	if currentWeek == news.CurrentLikesWeek {
		updateIncFields["weekLikes"] = 1
	} else {
		updateSetFields["weekLikes"] = 1
		updateSetFields["currentLikesWeek"] = currentWeek
	}

	currentMonth := utils.UTCNowBeginningOfMonth()
	if currentMonth == news.CurrentLikesMonth {
		updateIncFields["monthLikes"] = 1
	} else {
		updateSetFields["monthLikes"] = 1
		updateSetFields["currentLikesMonth"] = currentMonth
	}

	update["$inc"] = updateIncFields
	update["$set"] = updateSetFields

	updateFieldLikeBy := primitive.M{
		"likedBy": likeBy.Username,
	}
	update["$addToSet"] = updateFieldLikeBy

	err := newsDAO.UpdateByPKey(ctx, news.NewsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - IncreaseLikes): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}

func (newsDAO *NewsDAO) Rate(ctx context.Context, newsId string, point int32, rateBy *model.User) (*model.News, error) {
	mutexName := fmt.Sprintf("%v_%v", newsDAO.CollectionName, newsId)

	redLock := newsDAO.CreateRedlock(ctx, mutexName, newsDAO.CacheLockTTL)

	// Obtain a lock for our given mutex. After this is successful, no one else
	//	// can obtain the same lock (the same mutex name) until we unlock it.
	log.Println("try acquire Lock", mutexName)
	err := redLock.Lock()
	// Release the lock so other processes or threads can obtain a lock.
	defer func() {
		//redLock.Unlock()
		ok, err := redLock.Unlock()
		log.Println("release Lock", mutexName, ok, err)
	}()

	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - Rate): failed executing redLock.Lock -> %w", err)
	}

	news, err := newsDAO.FindByNewsId(ctx, newsId)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - Rate): failed executing FindByNewsId -> %w", err)
	}

	result := &model.News{}

	accumulateRatingPoint := news.AccumulateRatingPoint + int64(point)
	ratingCount := news.RatingCount + 1
	var rating float64 = float64(accumulateRatingPoint) / float64(ratingCount)

	update := primitive.M{}

	//updateFieldsInc := primitive.M{}
	//updateFieldsInc["accumulateRatingPoint"] = point
	//updateFieldsInc["ratingCount"] = 1
	//update["$inc"] = updateFieldsInc

	updateSetFields := primitive.M{}
	updateSetFields["accumulateRatingPoint"] = accumulateRatingPoint
	updateSetFields["ratingCount"] = ratingCount
	updateSetFields["rating"] = rating
	update["$set"] = updateSetFields

	updateFieldRatedBy := primitive.M{
		"ratedBy": rateBy.Username,
	}
	update["$addToSet"] = updateFieldRatedBy

	err = newsDAO.UpdateByPKey(ctx, newsId, update, []interface{}{}, false, result)
	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - Rate): failed executing UpdateByPKey -> %w", err)
	}

	return result, nil
}
