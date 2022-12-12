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

const (
	User_FIELD_EMAIL = "email"
)

type UserDAO struct {
	DAO
}

var singletonUserDAO *UserDAO
var onceUserDAO sync.Once

func GetUserDAO() *UserDAO {
	onceUserDAO.Do(func() {
		fmt.Println("Init UserDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		userDAO := UserDAO{}
		userDAO.Init(mongoCtx, &db.MongoDb)

		singletonUserDAO = &userDAO
	})
	return singletonUserDAO
}

func (userDAO *UserDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "users"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	userDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{User_FIELD_EMAIL}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (userDAO *UserDAO) Save(ctx context.Context, user *model.User) (*model.User, error) {
	result := &model.User{}
	err := userDAO.InsertOrUpdate(ctx, user, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - Save): failed executing Save -> %w", err)
	}

	err = userDAO.SaveToCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - Save): failed executing SaveToCache -> %w", err)
	}

	return result, nil
}

func (userDAO *UserDAO) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	result := &model.User{}

	err := userDAO.GetFromCache(ctx, utils.CacheKey(userDAO.CollectionName, User_FIELD_EMAIL, email), result)
	if err == nil {
		log.Println("FindByEmail found in cache OK")
		return result, nil
	}

	log.Println("FindByEmail not found in cache -> find in db")
	err = userDAO.FindByField(ctx, User_FIELD_EMAIL, email, result)

	if err != nil {
		return nil, fmt.Errorf("(UserDAO - FindByEmail): failed executing FindByField -> %w", err)
	}

	err = userDAO.SaveToCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - FindByEmail): failed executing SaveToCache -> %w", err)
	}

	return result, nil
}

func (userDAO *UserDAO) SearchByUserName(ctx context.Context, username string) ([]*model.User, error) {
	results := []*model.User{}

	//{username: {$regex: ".*GdK.*", $options: "i"}}

	pattern := fmt.Sprintf(".*%v.*", username)
	search := bson.M{
		"$regex":   pattern,
		"$options": "i",
	}
	filter := bson.M{PKEY_NAME: search}

	findOptions := options.Find()
	findOptions.SetSort(primitive.M{"createdAt": 1})
	findOptions.SetLimit(20)

	err := userDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(UserDAO - SearchByUserName): failed executing SearchByField -> %w", err)
	}

	return results, nil
}

func (userDAO *UserDAO) FindByListUserName(ctx context.Context, usernames []string) ([]*model.User, error) {
	results := []*model.User{}

	//{pkey: {$in: ["xxx","yyy","zzz"]}}

	filter := bson.M{PKEY_NAME: primitive.M{
		"$in": usernames,
	}}

	findOptions := options.Find()
	findOptions.SetSort(primitive.M{"createdAt": 1})

	err := userDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(UserDAO - FindByListUserName): failed executing SearchByField -> %w", err)
	}

	return results, nil
}

func (userDAO *UserDAO) FindByUserName(ctx context.Context, username string) (*model.User, error) {
	result := &model.User{}

	err := userDAO.GetFromCache(ctx, utils.CacheKey(userDAO.CollectionName, PKEY_NAME, username), result)
	if err == nil {
		return result, nil
	}

	err = userDAO.FindByPKey(ctx, username, result)

	if err != nil {
		return nil, fmt.Errorf("(UserDAO - FindByUserName): failed executing FindByField -> %w", err)
	}

	err = userDAO.SaveToCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - FindByUserName): failed executing SaveToCache -> %w", err)
	}

	return result, nil
}

func (userDAO *UserDAO) DeleteByUserName(ctx context.Context, username string) (*model.User, error) {
	result := &model.User{}
	err := userDAO.DeleteByPKey(ctx, username, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - DeleteByUserName): failed executing DeleteOneByField -> %w", err)
	}

	err = userDAO.DeleteFromCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - DeleteByUserName): failed executing DeleteFromCache -> %w", err)
	}
	return result, nil
}

func (userDAO *UserDAO) DeleteByEmail(ctx context.Context, email string) (*model.User, error) {
	result := &model.User{}
	err := userDAO.DeleteOneByField(ctx, User_FIELD_EMAIL, email, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - DeleteByEmail): failed executing DeleteOneByField -> %w", err)
	}

	err = userDAO.DeleteFromCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - DeleteByEmail): failed executing DeleteFromCache -> %w", err)
	}
	return result, nil
}

func (userDAO *UserDAO) UpdateByUserName(ctx context.Context, username string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.User, error) {
	result := &model.User{}

	err := userDAO.UpdateByPKey(ctx, username, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - UpdateByUserName): failed executing UpdateOneByField -> %w", err)
	}

	err = userDAO.SaveToCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - UpdateByUserName): failed executing SaveToCache -> %w", err)
	}
	return result, nil
}

func (userDAO *UserDAO) UpdateByEmail(ctx context.Context, email string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.User, error) {
	result := &model.User{}

	err := userDAO.UpdateOneByField(ctx, User_FIELD_EMAIL, email, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - UpdateByEmail): failed executing UpdateOneByField -> %w", err)
	}

	err = userDAO.SaveToCache(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("(UserDAO - UpdateByEmail): failed executing SaveToCache -> %w", err)
	}
	return result, nil
}

func (userDAO *UserDAO) SaveToCache(ctx context.Context, user *model.User) error {
	err := userDAO.SetToCache(ctx, utils.CacheKey(userDAO.CollectionName, PKEY_NAME, user.Username), user, userDAO.CacheTTL)
	if err != nil {
		return fmt.Errorf("(UserDAO - SaveToCache): failed executing SetToCache PKEY_NAME -> %w", err)
	}

	if len(user.Email) > 0 {
		err = userDAO.SetToCache(ctx, utils.CacheKey(userDAO.CollectionName, User_FIELD_EMAIL, user.Email), user, userDAO.CacheTTL)
		if err != nil {
			return fmt.Errorf("(UserDAO - SaveToCache): failed executing SetToCache User_FIELD_EMAIL -> %w", err)
		}
	}

	return nil
}

func (userDAO *UserDAO) DeleteFromCache(ctx context.Context, user *model.User) error {
	keys := []string{
		utils.CacheKey(userDAO.CollectionName, PKEY_NAME, user.Username),
		utils.CacheKey(userDAO.CollectionName, User_FIELD_EMAIL, user.Email),
	}
	err := userDAO.DelFromCache(ctx, keys)
	if err != nil {
		return fmt.Errorf("(UserDAO - DeleteFromCache): failed executing DelFromCache -> %w", err)
	}
	return nil
}
