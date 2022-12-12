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

type FileInfoDAO struct {
	DAO
}

var singletonFileInfoDAO *FileInfoDAO
var onceFileInfoDAO sync.Once

func GetFileInfoDAO() *FileInfoDAO {
	onceFileInfoDAO.Do(func() {
		fmt.Println("Init S3FileInfoDAO...")

		db := GetDataBase()
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancelMongo()

		fileInfoDAO := FileInfoDAO{}
		fileInfoDAO.Init(mongoCtx, &db.MongoDb)

		singletonFileInfoDAO = &fileInfoDAO
	})
	return singletonFileInfoDAO
}

func (fileInfoDAO *FileInfoDAO) Init(ctx context.Context, db *mongo.Database) {
	COLLECTION_NAME := "fileInfos"
	CACHE_TTL := 10 * time.Minute
	CACHE_LOCK_TTL := 30 * time.Second
	fileInfoDAO.InitDAO(ctx, db, COLLECTION_NAME, []string{}, CACHE_TTL, CACHE_LOCK_TTL)
}

func (fileInfoDAO *FileInfoDAO) Save(ctx context.Context, fileInfo *model.FileInfo) (*model.FileInfo, error) {
	result := &model.FileInfo{}
	err := fileInfoDAO.InsertOrUpdate(ctx, fileInfo, result)
	if err != nil {
		return nil, fmt.Errorf("(FileInfoDAO - Save): failed executing Save -> %w", err)
	}

	return result, nil
}

func (fileInfoDAO *FileInfoDAO) FindByFileId(ctx context.Context, fileId string) (*model.FileInfo, error) {
	result := &model.FileInfo{}

	err := fileInfoDAO.FindByPKey(ctx, fileId, result)
	if err != nil {
		return nil, fmt.Errorf("(FileInfoDAO - FindByFileId): failed executing FindByField -> %w", err)
	}

	return result, nil
}

func (fileInfoDAO *FileInfoDAO) FindByNewsId(ctx context.Context, newsId string) ([]*model.FileInfo, error) {
	results := []*model.FileInfo{}

	filter := primitive.M{}
	filter["newsId"] = newsId

	findOptions := options.Find()

	err := fileInfoDAO.FindAll(ctx, filter, findOptions, &results)
	if err != nil {
		return nil, fmt.Errorf("(FileInfoDAO - FindByNewsId): failed executing FindByField -> %w", err)
	}

	return results, nil
}

func (fileInfoDAO *FileInfoDAO) FindByOnDemandMediaMainFileId(ctx context.Context, onDemandMediaMainFileId string) ([]*model.FileInfo, error) {
	results := []*model.FileInfo{}

	filter := primitive.M{"onDemandMediaMainFileId": onDemandMediaMainFileId}

	sort := primitive.M{"createdAt": -1}
	findOptions := options.Find()
	findOptions.SetSort(sort)

	err := fileInfoDAO.FindAll(ctx, filter, findOptions, &results)

	if err != nil {
		return nil, fmt.Errorf("(NewsDAO - FindByOnDemandMediaMainFileId): failed executing FindAll -> %w", err)
	}

	return results, nil
}

func (fileInfoDAO *FileInfoDAO) DeleteByFileId(ctx context.Context, fileId string) (*model.FileInfo, error) {
	result := &model.FileInfo{}
	err := fileInfoDAO.DeleteByPKey(ctx, fileId, result)
	if err != nil {
		return nil, fmt.Errorf("(FileInfoDAO - DeleteByFileId): failed executing DeleteOneByField -> %w", err)
	}
	return result, nil
}

func (fileInfoDAO *FileInfoDAO) DeleteByOnDemandMediaMainFileId(ctx context.Context, onDemandMediaMainFileId string) error {
	err := fileInfoDAO.DeleteManyByField(ctx, "onDemandMediaMainFileId", onDemandMediaMainFileId)
	if err != nil {
		return fmt.Errorf("(FileInfoDAO - DeleteByOnDemandMediaMainFileId): failed executing DeleteManyByField -> %w", err)
	}
	return nil
}

func (fileInfoDAO *FileInfoDAO) DeleteByNewsId(ctx context.Context, newsId string) error {
	err := fileInfoDAO.DeleteManyByField(ctx, "newsId", newsId)
	if err != nil {
		return fmt.Errorf("(FileInfoDAO - DeleteByNewsId): failed executing DeleteManyByField -> %w", err)
	}
	return nil
}

func (fileInfoDAO *FileInfoDAO) CountByFileNameAndNewsId(ctx context.Context, fileName string, newsId string) (int64, error) {
	filter := bson.M{
		"fileName": fileName,
		"newsId":   newsId,
	}

	result, err := fileInfoDAO.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("(NewsDAO - CountByCatId): failed executing CountDocuments -> %w", err)
	}

	return result, nil
}
