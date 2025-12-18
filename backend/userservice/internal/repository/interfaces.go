package repository

import (
	"context"
	"userservice/internal/domain"
)

// UserRepository - основной интерфейс репозитория пользователей
type UserRepository interface {
	// Основные CRUD операции
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error

	// Поиск и фильтрация
	List(ctx context.Context, filter *domain.UserFilter) ([]*domain.User, int64, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	Exists(ctx context.Context, email, username string) (bool, error)

	// Операции с баном
	Ban(ctx context.Context, userID string, banInfo *domain.BanInfo) error
	Unban(ctx context.Context, userID string) error

	// Операции с подписками
	UpdateSubscription(ctx context.Context, userID string, subscription *domain.SubscriptionInfo) error
	CancelSubscription(ctx context.Context, userID string, reason string, immediate bool) error

	// Операции с последним входом
	UpdateLastLogin(ctx context.Context, userID string) error

	// Health check
	Ping(ctx context.Context) error
}

// AuditRepository - репозиторий для аудита и метаданных (в MongoDB)
type AuditRepository interface {
	// Операции с метаданными
	SaveMetadata(ctx context.Context, userID string, metadata map[string]string) error
	GetMetadata(ctx context.Context, userID string) (map[string]string, error)
	UpdateMetadata(ctx context.Context, userID string, metadata map[string]string) error
	DeleteMetadata(ctx context.Context, userID string, keys []string) error

	// Логирование действий
	LogActivity(ctx context.Context, activity *domain.UserActivity) error
	GetUserActivities(ctx context.Context, userID string, limit int) ([]*domain.UserActivity, error)

	// История подписок
	LogSubscriptionChange(ctx context.Context, entry *domain.SubscriptionHistoryEntry) error
	GetSubscriptionHistory(ctx context.Context, userID string) ([]*domain.SubscriptionHistoryEntry, error)

	// История банов
	LogBanChange(ctx context.Context, userID string, action string, details map[string]interface{}) error
	GetBanHistory(ctx context.Context, userID string) ([]map[string]interface{}, error)

	// Health check
	Ping(ctx context.Context) error
}
