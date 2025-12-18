package mongodb

import (
	"context"
	"fmt"
	"time"

	"userservice/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoUserRepository - репозиторий для MongoDB (аудит и метаданные)
type MongoUserRepository struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewMongoUserRepository создает новый репозиторий MongoDB
func NewMongoUserRepository(client *mongo.Client, dbName string) *MongoUserRepository {
	return &MongoUserRepository{
		client: client,
		db:     client.Database(dbName),
	}
}

// Структуры для MongoDB

// UserMetadataDocument - документ метаданных пользователя
type UserMetadataDocument struct {
	ID        string            `bson:"_id,omitempty"`
	UserID    string            `bson:"user_id"`
	Metadata  map[string]string `bson:"metadata"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
}

// UserActivityDocument - документ активности пользователя
type UserActivityDocument struct {
	ID           string         `bson:"_id,omitempty"`
	UserID       string         `bson:"user_id"`
	ActivityType string         `bson:"activity_type"`
	IPAddress    string         `bson:"ip_address,omitempty"`
	UserAgent    string         `bson:"user_agent,omitempty"`
	DeviceID     string         `bson:"device_id,omitempty"`
	Details      map[string]any `bson:"details,omitempty"`
	CreatedAt    time.Time      `bson:"created_at"`
}

// SubscriptionHistoryDocument - документ истории подписок
type SubscriptionHistoryDocument struct {
	ID        string         `bson:"_id,omitempty"`
	UserID    string         `bson:"user_id"`
	OldLevel  string         `bson:"old_level,omitempty"`
	NewLevel  string         `bson:"new_level"`
	OldStatus string         `bson:"old_status,omitempty"`
	NewStatus string         `bson:"new_status"`
	Reason    string         `bson:"reason,omitempty"`
	ChangedBy string         `bson:"changed_by"`
	ChangedAt time.Time      `bson:"changed_at"`
	Metadata  map[string]any `bson:"metadata,omitempty"`
}

// BanHistoryDocument - документ истории банов
type BanHistoryDocument struct {
	ID        string                 `bson:"_id,omitempty"`
	UserID    string                 `bson:"user_id"`
	Action    string                 `bson:"action"` // "ban", "unban"
	Reason    string                 `bson:"reason,omitempty"`
	Duration  *time.Duration         `bson:"duration,omitempty"`
	BannedBy  string                 `bson:"banned_by"`
	Details   map[string]interface{} `bson:"details,omitempty"`
	CreatedAt time.Time              `bson:"created_at"`
}

// SaveMetadata сохраняет метаданные пользователя
func (r *MongoUserRepository) SaveMetadata(ctx context.Context, userID string, metadata map[string]string) error {
	collection := r.db.Collection("user_metadata")

	filter := bson.D{{Key: "user_id", Value: userID}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "metadata", Value: metadata},
			{Key: "updated_at", Value: time.Now()},
		}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "user_id", Value: userID},
			{Key: "created_at", Value: time.Now()},
		}},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// GetMetadata получает метаданные пользователя
func (r *MongoUserRepository) GetMetadata(ctx context.Context, userID string) (map[string]string, error) {
	collection := r.db.Collection("user_metadata")

	var doc UserMetadataDocument
	filter := bson.D{{Key: "user_id", Value: userID}}

	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return doc.Metadata, nil
}

// UpdateMetadata обновляет метаданные пользователя
func (r *MongoUserRepository) UpdateMetadata(ctx context.Context, userID string, metadata map[string]string) error {
	// Получаем текущие метаданные
	currentMetadata, err := r.GetMetadata(ctx, userID)
	if err != nil {
		return err
	}

	// Обновляем метаданные
	for k, v := range metadata {
		currentMetadata[k] = v
	}

	// Сохраняем обновленные метаданные
	return r.SaveMetadata(ctx, userID, currentMetadata)
}

// DeleteMetadata удаляет ключи из метаданных
func (r *MongoUserRepository) DeleteMetadata(ctx context.Context, userID string, keys []string) error {
	collection := r.db.Collection("user_metadata")

	// Создаем update для удаления ключей
	unsetFields := bson.D{}
	for _, key := range keys {
		unsetFields = append(unsetFields, bson.E{Key: "metadata." + key, Value: ""})
	}

	filter := bson.D{{Key: "user_id", Value: userID}}
	_, err := collection.UpdateOne(ctx, filter, bson.D{{Key: "$unset", Value: unsetFields}})
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// LogActivity логирует активность пользователя
func (r *MongoUserRepository) LogActivity(ctx context.Context, activity *domain.UserActivity) error {
	collection := r.db.Collection("user_activities")

	doc := &UserActivityDocument{
		ID:           activity.ID,
		UserID:       activity.UserID,
		ActivityType: string(activity.ActivityType),
		IPAddress:    activity.IPAddress,
		UserAgent:    activity.UserAgent,
		DeviceID:     activity.DeviceID,
		Details:      activity.Details,
		CreatedAt:    activity.CreatedAt,
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// GetUserActivities получает активность пользователя
func (r *MongoUserRepository) GetUserActivities(ctx context.Context, userID string, limit int) ([]*domain.UserActivity, error) {
	collection := r.db.Collection("user_activities")

	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	filter := bson.D{{Key: "user_id", Value: userID}}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find activities: %w", err)
	}
	defer cursor.Close(ctx)

	var activities []*domain.UserActivity
	for cursor.Next(ctx) {
		var doc UserActivityDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		activity := &domain.UserActivity{
			ID:           doc.ID,
			UserID:       doc.UserID,
			ActivityType: domain.ActivityType(doc.ActivityType),
			IPAddress:    doc.IPAddress,
			UserAgent:    doc.UserAgent,
			DeviceID:     doc.DeviceID,
			Details:      doc.Details,
			CreatedAt:    doc.CreatedAt,
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// LogSubscriptionChange логирует изменение подписки
func (r *MongoUserRepository) LogSubscriptionChange(ctx context.Context, entry *domain.SubscriptionHistoryEntry) error {
	collection := r.db.Collection("subscription_history")

	doc := &SubscriptionHistoryDocument{
		ID:        entry.ID,
		UserID:    entry.UserID,
		OldLevel:  string(entry.OldLevel),
		NewLevel:  string(entry.NewLevel),
		OldStatus: string(entry.OldStatus),
		NewStatus: string(entry.NewStatus),
		Reason:    entry.Reason,
		ChangedBy: entry.ChangedBy,
		ChangedAt: entry.ChangedAt,
		Metadata:  entry.Metadata,
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to log subscription change: %w", err)
	}

	return nil
}

// GetSubscriptionHistory получает историю подписок пользователя
func (r *MongoUserRepository) GetSubscriptionHistory(ctx context.Context, userID string) ([]*domain.SubscriptionHistoryEntry, error) {
	collection := r.db.Collection("subscription_history")

	filter := bson.D{{Key: "user_id", Value: userID}}
	opts := options.Find().SetSort(bson.D{{Key: "changed_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find subscription history: %w", err)
	}
	defer cursor.Close(ctx)

	var entries []*domain.SubscriptionHistoryEntry
	for cursor.Next(ctx) {
		var doc SubscriptionHistoryDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		entry := &domain.SubscriptionHistoryEntry{
			ID:        doc.ID,
			UserID:    doc.UserID,
			OldLevel:  domain.SubscriptionLevel(doc.OldLevel),
			NewLevel:  domain.SubscriptionLevel(doc.NewLevel),
			OldStatus: domain.SubscriptionStatus(doc.OldStatus),
			NewStatus: domain.SubscriptionStatus(doc.NewStatus),
			Reason:    doc.Reason,
			ChangedBy: doc.ChangedBy,
			ChangedAt: doc.ChangedAt,
			Metadata:  doc.Metadata,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// LogBanChange логирует изменение бана
func (r *MongoUserRepository) LogBanChange(ctx context.Context, userID string, action string, details map[string]interface{}) error {
	collection := r.db.Collection("ban_history")

	doc := &BanHistoryDocument{
		UserID:    userID,
		Action:    action,
		Reason:    getString(details, "reason"),
		BannedBy:  getString(details, "banned_by"),
		Details:   details,
		CreatedAt: time.Now(),
	}

	// Извлекаем duration если есть
	if dur, ok := details["duration"].(time.Duration); ok {
		doc.Duration = &dur
	}

	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to log ban change: %w", err)
	}

	return nil
}

// GetBanHistory получает историю банов пользователя
func (r *MongoUserRepository) GetBanHistory(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	collection := r.db.Collection("ban_history")

	filter := bson.D{{Key: "user_id", Value: userID}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find ban history: %w", err)
	}
	defer cursor.Close(ctx)

	var history []map[string]interface{}
	for cursor.Next(ctx) {
		var doc BanHistoryDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		entry := map[string]interface{}{
			"action":     doc.Action,
			"reason":     doc.Reason,
			"banned_by":  doc.BannedBy,
			"duration":   doc.Duration,
			"details":    doc.Details,
			"created_at": doc.CreatedAt,
		}

		history = append(history, entry)
	}

	return history, nil
}

// Ping проверяет соединение с MongoDB
func (r *MongoUserRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx, nil)
}

// Вспомогательная функция для получения строки из map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
