package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoConfig - конфигурация MongoDB
type MongoConfig struct {
	URI         string
	Database    string
	MaxPoolSize uint64
	MinPoolSize uint64
	Timeout     time.Duration
}

// ConnectMongoDB подключается к MongoDB
func ConnectMongoDB(cfg MongoConfig) (*mongo.Client, error) {
	// Настройки клиента
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize)

	// Подключение
	_, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Проверка соединения
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()

	if err := client.Ping(ctxPing, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	return client, nil
}

// CreateIndexes создает индексы для коллекций MongoDB
func CreateIndexes(client *mongo.Client, dbName string) error {
	db := client.Database(dbName)

	// Индексы для user_metadata
	userMetadataIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"user_id": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"created_at": -1},
		},
	}

	// Индексы для user_activities
	userActivitiesIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
		},
		{
			Keys: map[string]interface{}{"activity_type": 1},
		},
	}

	// Индексы для subscription_history
	subscriptionHistoryIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "changed_at": -1},
		},
	}

	// Индексы для ban_history
	banHistoryIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
		},
	}

	// Создаем индексы
	ctx := context.Background()

	collections := map[string][]mongo.IndexModel{
		"user_metadata":        userMetadataIndexes,
		"user_activities":      userActivitiesIndexes,
		"subscription_history": subscriptionHistoryIndexes,
		"ban_history":          banHistoryIndexes,
	}

	for collectionName, indexes := range collections {
		collection := db.Collection(collectionName)
		_, err := collection.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			return fmt.Errorf("failed to create indexes for %s: %w", collectionName, err)
		}
	}

	return nil
}
