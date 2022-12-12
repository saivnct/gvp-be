package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/go-redsync/redsync/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"strings"
	"time"
)

type DAO struct {
	Collection     mongo.Collection
	CollectionName string
	CacheTTL       time.Duration
	CacheLockTTL   time.Duration
}

const PKEY_NAME = "pkey"

func (d *DAO) InitDAO(ctx context.Context, db *mongo.Database,
	colName string, uniqueFields []string, cacheTTL time.Duration, cacheLockTTL time.Duration) {

	ufileds := []string{PKEY_NAME}
	ufileds = append(ufileds, uniqueFields...)

	d.LoadCollection(ctx, db, colName, ufileds)
	d.CollectionName = colName
	d.CacheTTL = cacheTTL
	d.CacheLockTTL = cacheLockTTL
}

func (d *DAO) LoadCollection(ctx context.Context, db *mongo.Database, colName string, uniqueFields []string) {
	d.Collection = *db.Collection(colName)
	for _, uniqueField := range uniqueFields {
		_, err := d.Collection.Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys:    bson.D{{Key: uniqueField, Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		)

		if err != nil {
			log.Fatal(err)
		}

	}
}

func (d *DAO) InsertOrUpdate(ctx context.Context, item interface{}, result interface{}) error {
	validate := validator.New()
	err := validate.Struct(item)
	if err != nil {
		return fmt.Errorf("(DAO - Save): failed validating object -> %w", err)
	}

	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	update := primitive.M{}
	updateFields := bson.M{}

	var pkeyValue = ""

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldValue := field.Interface()
		if field.Kind() == reflect.Ptr {
			fieldValue = field.Elem().Interface()
		}

		t := val.Type().Field(i)
		//fieldName := t.Name

		bsonName := ""
		switch bsonTag := t.Tag.Get("bson"); bsonTag {
		case "-":
		case "":
			bsonName = ""
		default:
			parts := strings.Split(bsonTag, ",")
			bsonName = parts[0]
		}

		if bsonName == PKEY_NAME && field.Kind() == reflect.String {
			pkeyValue = field.String()
		}

		if bsonName == "_id" {
			id, ok := fieldValue.(primitive.ObjectID)
			if ok && id.IsZero() {
				//Ignore zero ID
				continue
			}
		}

		if len(bsonName) > 0 {
			updateFields[bsonName] = fieldValue
		}

		update["$set"] = updateFields
		//fmt.Printf("fieldName: %s\tbsonName: %s\tfieldValue: %v\n", fieldName, bsonName, fieldValue)
	}

	//fmt.Printf("pkeyName: %s\tpkeyValue: %s\n", PKEY_NAME, pkeyValue)

	if len(pkeyValue) > 0 {
		return d.UpdateByPKey(ctx, pkeyValue, update, []interface{}{}, true, result)
	}

	return fmt.Errorf("(DAO - Save): Invalid pkey value")

}

func (d *DAO) InsertOne(ctx context.Context, item interface{}) (*primitive.ObjectID, error) {
	result, err := d.Collection.InsertOne(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("(DAO - InsertOne): failed executing db InsertOne -> %w", err)
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("(DAO - InsertOne): failed reading ObjectID")
	}

	return &oid, nil
}

func (d *DAO) FindById(ctx context.Context, id string, result interface{}) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("(DAO - FindById): failed converting ObjectID -> %w", err)
	}

	filter := bson.M{"_id": oid}
	err = d.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - FindById): failed executing db FindOne -> %w", err)
	}
	return nil
}

func (d *DAO) FindByOid(ctx context.Context, oid primitive.ObjectID, result interface{}) error {
	filter := bson.M{"_id": oid}
	err := d.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - FindByOid): failed executing db FindOne -> %w", err)
	}
	return nil
}

func (d *DAO) FindByPKey(ctx context.Context, pkeyValue string, result interface{}) error {
	filter := bson.M{PKEY_NAME: pkeyValue}
	err := d.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - FindByPKey): failed executing db FindOne -> %w", err)
	}
	return nil
}

func (d *DAO) FindByField(ctx context.Context, fieldName, fieldValue string, result interface{}) error {
	filter := bson.M{fieldName: fieldValue}

	err := d.Collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - FindByField): failed executing db FindOne -> %w", err)
	}
	return nil
}

func (d *DAO) UpdateById(ctx context.Context, id string, update interface{}, arrayFilter []interface{}, result interface{}) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("(DAO - UpdateById): failed converting ObjectID -> %w", err)
	}

	filter := bson.M{"_id": oid}
	err = d.Collection.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilter,
		}),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - UpdateById): failed executing db FindOneAndUpdate -> %w", err)
	}
	return nil
}

func (d *DAO) UpdateByOid(ctx context.Context, oid primitive.ObjectID, update interface{}, arrayFilter []interface{}, result interface{}) error {
	filter := bson.M{"_id": oid}
	err := d.Collection.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilter,
		}),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - UpdateByOid): failed executing db FindOneAndUpdate -> %w", err)
	}
	return nil
}

func (d *DAO) UpdateByPKey(ctx context.Context, pkeyValue string, update interface{}, arrayFilter []interface{}, upsert bool, result interface{}) error {
	filter := bson.M{PKEY_NAME: pkeyValue}
	err := d.Collection.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetUpsert(upsert),
		options.FindOneAndUpdate().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilter,
		}),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - UpdateByPKey): failed executing db FindOneAndUpdate -> %w", err)
	}
	return nil
}

func (d *DAO) UpdateOneByField(ctx context.Context, fieldName, fieldValue string, update interface{}, arrayFilter []interface{}, upsert bool, result interface{}) error {
	filter := bson.M{fieldName: fieldValue}

	err := d.Collection.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetUpsert(upsert),
		options.FindOneAndUpdate().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilter,
		}),
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - UpdateOneByField): failed executing db FindOneAndUpdate -> %w", err)
	}
	return nil
}

func (d *DAO) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (int64, error) {
	result, err := d.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("(DAO - UpdateMany): failed executing db UpdateMany -> %w", err)
	}
	return result.ModifiedCount, nil
}

func (d *DAO) FindAll(ctx context.Context, filter interface{}, findOptions *options.FindOptions, result interface{}) error {
	cursor, err := d.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		return fmt.Errorf("(DAO - FindAll): failed executing db Find -> %w", err)
	}

	defer cursor.Close(ctx) //when this function exit, this cursor will Close

	err = cursor.All(ctx, result)
	if err != nil {
		return fmt.Errorf("(DAO - FindAll): failed executing cursor.All -> %w", err)
	}
	return nil
}

func (d *DAO) FindWithPagination(ctx context.Context, filter interface{}, findOptions *options.FindOptions, page, pageSize int64, result interface{}) error {
	findOptions.SetSkip(page * pageSize)
	findOptions.SetLimit(pageSize)

	cursor, err := d.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		return fmt.Errorf("(DAO - FindWithPagination): failed executing db Find -> %w", err)
	}

	defer cursor.Close(ctx) //when this function exit, this cursor will Close

	err = cursor.All(ctx, result)
	if err != nil {
		return fmt.Errorf("(DAO - FindWithPagination): failed executing cursor.All -> %w", err)
	}
	return nil
}

func (d *DAO) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	total, err := d.Collection.CountDocuments(ctx, filter)

	if err != nil {
		return 0, fmt.Errorf("(DAO - CountDocuments): failed executing db CountDocuments -> %w", err)
	}

	return total, nil
}

func (d *DAO) DeleteById(ctx context.Context, id string, result interface{}) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteById): failed converting ObjectID -> %w", err)
	}

	filter := bson.M{"_id": oid}

	err = d.Collection.FindOneAndDelete(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteById): failed executing db FindOneAndDelete -> %w", err)
	}
	return nil
}

func (d *DAO) DeleteByPKey(ctx context.Context, pkeyValue string, result interface{}) error {
	filter := bson.M{PKEY_NAME: pkeyValue}

	err := d.Collection.FindOneAndDelete(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteByPKey): failed executing db FindOneAndDelete -> %w", err)
	}
	return nil
}

func (d *DAO) DeleteByOid(ctx context.Context, oid primitive.ObjectID, result interface{}) error {
	filter := bson.M{"_id": oid}

	err := d.Collection.FindOneAndDelete(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteByOid): failed executing db FindOneAndDelete -> %w", err)
	}
	return nil
}

func (d *DAO) DeleteOneByField(ctx context.Context, fieldName, fieldValue string, result interface{}) error {
	filter := bson.M{fieldName: fieldValue}

	err := d.Collection.FindOneAndDelete(ctx, filter).Decode(result)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteOneByField): failed executing db FindOneAndDelete -> %w", err)
	}
	return nil
}

func (d *DAO) DeleteManyByField(ctx context.Context, fieldName, fieldValue string) error {
	filter := bson.M{fieldName: fieldValue}

	_, err := d.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("(DAO - DeleteManyByField): failed executing db DeleteMany -> %w", err)
	}
	return nil
}

func (d *DAO) SetToCache(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	jsonString, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("(DAO - SetToCache): failed executing json.Marshal -> %w", err)
	}

	cache := GetCache()
	err = cache.RedisClient.Set(ctx, key, string(jsonString), expiration).Err()
	if err != nil {
		return fmt.Errorf("(DAO - SetToCache): failed executing cache Set -> %w", err)
	}
	return nil
}

func (d *DAO) GetFromCache(ctx context.Context, key string, result interface{}) error {
	cache := GetCache()
	productsCache, err := cache.RedisClient.Get(ctx, key).Bytes()

	if err != nil {
		return fmt.Errorf("(DAO - GetFromCache): failed executing cache Get -> %w", err)
	}

	err = json.Unmarshal(productsCache, result)
	if err != nil {
		return fmt.Errorf("(DAO - GetFromCache): failed executing json.Unmarshal -> %w", err)
	}
	return nil
}

func (d *DAO) DelFromCache(ctx context.Context, keys []string) error {
	cache := GetCache()
	err := cache.RedisClient.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("(DAO - DelFromCache): failed executing cache Del -> %w", err)
	}
	return nil
}

func (d *DAO) CreateRedlock(ctx context.Context, mutexName string, expiration time.Duration) *redsync.Mutex {
	cache := GetCache()
	options := []redsync.Option{redsync.WithExpiry(expiration), redsync.WithTries(1000)}
	mutex := cache.RedSync.NewMutex(mutexName, options...)
	return mutex
}

func (d *DAO) CreateRedlockNoRetry(ctx context.Context, mutexName string, expiration time.Duration) *redsync.Mutex {
	cache := GetCache()
	options := []redsync.Option{redsync.WithExpiry(expiration), redsync.WithTries(1)}
	mutex := cache.RedSync.NewMutex(mutexName, options...)
	return mutex
}
