package dao

import (
	"context"
	"fmt"
	"gbb.go/gvp/model"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type EmailLockRegDAO struct {
	DAO
}

var singletonPhoneLockRegDAO *EmailLockRegDAO
var oncePhoneLockRegDAO sync.Once

func GetEmailLockRegDAO() *EmailLockRegDAO {
	oncePhoneLockRegDAO.Do(func() {
		fmt.Println("Init EmailLockRegDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		emailLockRegDAO := EmailLockRegDAO{}
		emailLockRegDAO.Init(mongoCtx, &db.MongoDb)

		singletonPhoneLockRegDAO = &emailLockRegDAO
	})
	return singletonPhoneLockRegDAO
}

func (emailLockRegDAO *EmailLockRegDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "emailLockRegs"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	emailLockRegDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (emailLockRegDAO *EmailLockRegDAO) Save(ctx context.Context, emailLockReg *model.EmailLockReg) (*model.EmailLockReg, error) {
	result := &model.EmailLockReg{}
	err := emailLockRegDAO.InsertOrUpdate(ctx, emailLockReg, result)
	if err != nil {
		return nil, fmt.Errorf("(EmailLockRegDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (emailLockRegDAO *EmailLockRegDAO) FindByEmail(ctx context.Context, email string) (*model.EmailLockReg, error) {
	result := &model.EmailLockReg{}

	err := emailLockRegDAO.FindByPKey(ctx, email, result)
	if err != nil {
		return nil, fmt.Errorf("(EmailLockRegDAO - FindByEmail): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (emailLockRegDAO *EmailLockRegDAO) DeleteByEmail(ctx context.Context, email string) (*model.EmailLockReg, error) {
	result := &model.EmailLockReg{}
	err := emailLockRegDAO.DeleteByPKey(ctx, email, result)
	if err != nil {
		return nil, fmt.Errorf("(EmailLockRegDAO - DeleteByEmail): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (emailLockRegDAO *EmailLockRegDAO) UpdateByEmail(ctx context.Context, email string, update interface{}, arrayFilter []interface{}, upsert bool) (*model.EmailLockReg, error) {
	result := &model.EmailLockReg{}

	err := emailLockRegDAO.UpdateByPKey(ctx, email, update, arrayFilter, upsert, result)
	if err != nil {
		return nil, fmt.Errorf("(EmailLockRegDAO - UpdateByEmail): failed executing UpdateOneByField -> %w", err)
	}
	return result, nil
}
