package dao

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"sync"
	"time"
)

type DataBase struct {
	MongoClient mongo.Client
	MongoDb     mongo.Database
}

var singletonDb *DataBase
var onceDb sync.Once

func GetDataBase() *DataBase {
	onceDb.Do(func() {
		fmt.Println("Init DataBase...")

		//connect to mongodb
		mongoCtx, cancelMongo := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelMongo()

		//log.Println("Connecting", os.Getenv("MONGO_CONNECTION_STR"))

		mongoClient, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION_STR")))
		if err != nil {
			log.Fatalf("Failed to connect to mongodb: %v", err)
		}

		// test mongo connection
		err = mongoClient.Ping(mongoCtx, nil)
		if err != nil {
			log.Fatalf("Failed to connect to mongodb: %v", err)
		}

		db := mongoClient.Database(os.Getenv("MONGO_DB"))

		singletonDb = &DataBase{
			MongoClient: *mongoClient,
			MongoDb:     *db,
		}
	})
	return singletonDb
}
