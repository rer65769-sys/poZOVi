package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User - основная доменная модель пользователя
type User struct {
	ID           string
	ServiceEmail string
	Name         string
	Password     string // Уже хешированный пароль
	Email        string
	Phone        string
	Status       UserStatus
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	BanInfo      *BanInfo
	Subscription *SubscriptionInfo
	Metadata     map[string]string
}

// BanInfo - информация о бане пользователя
type BanInfo struct {
	IsBanned    bool
	BannedAt    time.Time
	BannedUntil *time.Time // nil для перманентного бана
	Reason      string
	BannedBy    string // ID администратора
}

// SubscriptionInfo - информация о подписке пользователя
type SubscriptionInfo struct {
	Status            SubscriptionStatus
	Level             SubscriptionLevel
	SubscriptionStart time.Time
	SubscriptionEnd   *time.Time
	TrialEnd          *time.Time
	SubscriptionID    string // ID во внешней системе (Stripe и т.д.)
	PaymentMethod     string
	AutoRenew         bool
	NextBillingDate   *time.Time
	Amount            float64
	Currency          string
	CanceledAt        *time.Time
	CancelReason      string
	GracePeriodEnd    *time.Time
	Features          []string
}

// UserActivity - активность пользователя (для аудита в MongoDB)
type UserActivity struct {
	ID           string
	UserID       string
	ActivityType ActivityType
	IPAddress    string
	UserAgent    string
	DeviceID     string
	Details      map[string]interface{} // Гибкое поле для деталей
	CreatedAt    time.Time
}

// SubscriptionHistoryEntry - история изменений подписки
type SubscriptionHistoryEntry struct {
	ID        string
	UserID    string
	OldLevel  SubscriptionLevel
	NewLevel  SubscriptionLevel
	OldStatus SubscriptionStatus
	NewStatus SubscriptionStatus
	Reason    string
	ChangedBy string // user_id или system
	ChangedAt time.Time
	Metadata  map[string]interface{} // Дополнительные данные
}

// UserStatus - статусы пользователя
type UserStatus string

const (
	UserStatusUnspecified       UserStatus = "UNSPECIFIED"
	UserStatusActive            UserStatus = "ACTIVE"
	UserStatusInactive          UserStatus = "INACTIVE"
	UserStatusSuspended         UserStatus = "SUSPENDED"
	UserStatusPending           UserStatus = "PENDING"
	UserStatusDeleted           UserStatus = "DELETED"
	UserStatusBannedPermanently UserStatus = "BANNED_PERMANENTLY"
	UserStatusBannedTemporarily UserStatus = "BANNED_TEMPORARILY"
	UserStatusBannedByAdmin     UserStatus = "BANNED_BY_ADMIN"
	UserStatusBannedBySystem    UserStatus = "BANNED_BY_SYSTEM"
	UserStatusBannedForSpam     UserStatus = "BANNED_FOR_SPAM"
	UserStatusBannedForAbuse    UserStatus = "BANNED_FOR_ABUSE"
	UserStatusBannedForFraud    UserStatus = "BANNED_FOR_FRAUD"
)

// UserRole - роли пользователя
type UserRole string

const (
	UserRoleUnspecified UserRole = "UNSPECIFIED"
	UserRoleUser        UserRole = "USER"
	UserRoleModerator   UserRole = "MODERATOR"
	UserRoleAdmin       UserRole = "ADMIN"
	UserRoleSuperAdmin  UserRole = "SUPER_ADMIN"
	UserRoleBannedUser  UserRole = "BANNED_USER"
)

// SubscriptionStatus - статусы подписки
type SubscriptionStatus string

const (
	SubscriptionStatusUnspecified SubscriptionStatus = "UNSPECIFIED"
	SubscriptionStatusInactive    SubscriptionStatus = "INACTIVE"
	SubscriptionStatusTrial       SubscriptionStatus = "TRIAL"
	SubscriptionStatusActive      SubscriptionStatus = "ACTIVE"
	SubscriptionStatusPastDue     SubscriptionStatus = "PAST_DUE"
	SubscriptionStatusCanceled    SubscriptionStatus = "CANCELED"
	SubscriptionStatusExpired     SubscriptionStatus = "EXPIRED"
	SubscriptionStatusPaused      SubscriptionStatus = "PAUSED"
	SubscriptionStatusPending     SubscriptionStatus = "PENDING"
	SubscriptionStatusGracePeriod SubscriptionStatus = "GRACE_PERIOD"
	SubscriptionStatusUpgrading   SubscriptionStatus = "UPGRADING"
	SubscriptionStatusDowngrading SubscriptionStatus = "DOWNGRADING"
)

// SubscriptionLevel - уровни подписки
type SubscriptionLevel string

const (
	SubscriptionLevelUnspecified SubscriptionLevel = "UNSPECIFIED"
	SubscriptionLevelFree        SubscriptionLevel = "FREE"
	SubscriptionLevelBasic       SubscriptionLevel = "BASIC"
	SubscriptionLevelStandard    SubscriptionLevel = "STANDARD"
	SubscriptionLevelPro         SubscriptionLevel = "PRO"
	SubscriptionLevelPremium     SubscriptionLevel = "PREMIUM"
	SubscriptionLevelEnterprise  SubscriptionLevel = "ENTERPRISE"
	SubscriptionLevelLifetime    SubscriptionLevel = "LIFETIME"
	SubscriptionLevelStarter     SubscriptionLevel = "STARTER"
	SubscriptionLevelBusiness    SubscriptionLevel = "BUSINESS"
	SubscriptionLevelUltimate    SubscriptionLevel = "ULTIMATE"
)

// ActivityType - типы активности пользователя
type ActivityType string

const (
	ActivityTypeUnspecified       ActivityType = "UNSPECIFIED"
	ActivityTypeLogin             ActivityType = "LOGIN"
	ActivityTypeLogout            ActivityType = "LOGOUT"
	ActivityTypePasswordChange    ActivityType = "PASSWORD_CHANGE"
	ActivityTypeProfileUpdate     ActivityType = "PROFILE_UPDATE"
	ActivityTypeEmailVerification ActivityType = "EMAIL_VERIFICATION"
	ActivityTypeSubscriptionStart ActivityType = "SUBSCRIPTION_START"
	ActivityTypeSubscriptionEnd   ActivityType = "SUBSCRIPTION_END"
	ActivityTypeBan               ActivityType = "BAN"
	ActivityTypeUnban             ActivityType = "UNBAN"
)

// Domain ошибки
// var (
// 	ErrUserNotFound         = errors.New("user not found")
// 	ErrUserAlreadyExists    = errors.New("user already exists")
// 	ErrInvalidCredentials   = errors.New("invalid credentials")
// 	ErrUserBanned           = errors.New("user is banned")
// 	ErrInvalidEmail         = errors.New("invalid email")
// 	ErrInvalidPassword      = errors.New("invalid password")
// 	ErrSubscriptionRequired = errors.New("subscription required")
// 	ErrPermissionDenied     = errors.New("permission denied")
// 	ErrInvalidToken         = errors.New("invalid token")
// )

// ===== Методы для User =====

// IsActive проверяет, активен ли пользователь
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsBanned проверяет, забанен ли пользователь
func (u *User) IsBanned() bool {
	if u.BanInfo == nil {
		return false
	}

	if !u.BanInfo.IsBanned {
		return false
	}

	// Проверяем временный бан
	if u.BanInfo.BannedUntil != nil && time.Now().After(*u.BanInfo.BannedUntil) {
		// Срок бана истек
		u.BanInfo.IsBanned = false
		return false
	}

	return true
}

// HasValidSubscription проверяет, есть ли активная подписка
func (u *User) HasValidSubscription() bool {
	if u.Subscription == nil {
		return false
	}

	now := time.Now()

	// Проверяем статус подписки
	activeStatuses := []SubscriptionStatus{
		SubscriptionStatusActive,
		SubscriptionStatusTrial,
		SubscriptionStatusGracePeriod,
	}

	for _, status := range activeStatuses {
		if u.Subscription.Status == status {
			// Проверяем срок действия
			if u.Subscription.SubscriptionEnd != nil && now.After(*u.Subscription.SubscriptionEnd) {
				return false
			}
			return true
		}
	}

	return false
}

// CanAccessFeature проверяет доступ к фиче по подписке
func (u *User) CanAccessFeature(feature string) bool {
	if !u.HasValidSubscription() {
		return false
	}

	// Проверяем, есть ли фича в списке доступных
	for _, f := range u.Subscription.Features {
		if f == feature {
			return true
		}
	}

	return false
}

// IsAdmin проверяет, является ли пользователь администратором
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleSuperAdmin || u.Role == UserRoleModerator
}

// Validate проверяет валидность пользователя
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}

	if u.Password == "" {
		return ErrInvalidPassword
	}

	if u.Name == "" {
		u.Name = u.Email // Устанавливаем имя по умолчанию
	}

	return nil
}

// ===== Методы для BanInfo =====

// Ban банит пользователя
func (b *BanInfo) Ban(reason string, bannedBy string, duration *time.Duration) {
	b.IsBanned = true
	b.BannedAt = time.Now()
	b.Reason = reason
	b.BannedBy = bannedBy

	if duration != nil {
		until := time.Now().Add(*duration)
		b.BannedUntil = &until
	} else {
		b.BannedUntil = nil // Перманентный бан
	}
}

// Unban разбанивает пользователя
func (b *BanInfo) Unban() {
	b.IsBanned = false
	b.BannedUntil = nil
}

// IsTemporaryBan проверяет, временный ли бан
func (b *BanInfo) IsTemporaryBan() bool {
	return b.BannedUntil != nil
}

// BanDurationRemaining возвращает оставшееся время бана
func (b *BanInfo) BanDurationRemaining() *time.Duration {
	if !b.IsBanned || b.BannedUntil == nil {
		return nil
	}

	duration := time.Until(*b.BannedUntil)
	if duration < 0 {
		return nil
	}

	return &duration
}

// ===== Методы для SubscriptionInfo =====

// Activate активирует подписку
func (s *SubscriptionInfo) Activate(level SubscriptionLevel, amount float64, currency string) {
	s.Status = SubscriptionStatusActive
	s.Level = level
	s.Amount = amount
	s.Currency = currency
	s.SubscriptionStart = time.Now()

	// Устанавливаем срок подписки (по умолчанию 1 месяц)
	end := time.Now().AddDate(0, 1, 0)
	s.SubscriptionEnd = &end
	s.NextBillingDate = &end
}

// Cancel отменяет подписку
func (s *SubscriptionInfo) Cancel(reason string, immediate bool) {
	s.Status = SubscriptionStatusCanceled
	s.CancelReason = reason
	now := time.Now()
	s.CanceledAt = &now

	if immediate {
		s.SubscriptionEnd = &now
	}

	s.AutoRenew = false
}

// IsTrial проверяет, пробный ли период
func (s *SubscriptionInfo) IsTrial() bool {
	return s.Status == SubscriptionStatusTrial
}

// HasTrialExpired проверяет, истек ли триал
func (s *SubscriptionInfo) HasTrialExpired() bool {
	if s.TrialEnd == nil {
		return false
	}

	return time.Now().After(*s.TrialEnd)
}

// DaysUntilExpiration возвращает дни до истечения подписки
func (s *SubscriptionInfo) DaysUntilExpiration() *int {
	if s.SubscriptionEnd == nil {
		return nil
	}

	days := int(s.SubscriptionEnd.Sub(time.Now()).Hours() / 24)
	if days < 0 {
		days = 0
	}

	return &days
}

// UpdateLevel обновляет уровень подписки
func (s *SubscriptionInfo) UpdateLevel(newLevel SubscriptionLevel, newAmount float64) {
	oldLevel := s.Level
	s.Level = newLevel
	s.Amount = newAmount

	if newLevel != oldLevel {
		// Обновляем статус при смене уровня
		if newLevel == SubscriptionLevelFree {
			s.Status = SubscriptionStatusInactive
		} else {
			s.Status = SubscriptionStatusActive
		}
	}
}

// ===== Методы для UserActivity =====

// NewUserActivity создает новую запись активности
func NewUserActivity(userID string, activityType ActivityType, ipAddress, userAgent, deviceID string) *UserActivity {
	return &UserActivity{
		ID:           generateUUID(),
		UserID:       userID,
		ActivityType: activityType,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceID:     deviceID,
		Details:      make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}
}

// AddDetail добавляет деталь в активность
func (a *UserActivity) AddDetail(key string, value interface{}) {
	if a.Details == nil {
		a.Details = make(map[string]interface{})
	}
	a.Details[key] = value
}

// ===== Методы для SubscriptionHistoryEntry =====

// NewSubscriptionHistoryEntry создает новую запись истории подписки
func NewSubscriptionHistoryEntry(userID string, oldLevel, newLevel SubscriptionLevel,
	oldStatus, newStatus SubscriptionStatus, reason, changedBy string) *SubscriptionHistoryEntry {
	return &SubscriptionHistoryEntry{
		ID:        generateUUID(),
		UserID:    userID,
		OldLevel:  oldLevel,
		NewLevel:  newLevel,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Reason:    reason,
		ChangedBy: changedBy,
		ChangedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata добавляет метаданные в запись истории
func (e *SubscriptionHistoryEntry) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// ===== Интерфейсы репозиториев =====

// UserRepository определяет интерфейс для работы с хранилищем пользователей
type UserRepository interface {
	// CRUD операции
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error

	// Поиск и фильтрация
	List(ctx context.Context, filter *UserFilter) ([]*User, int64, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Exists(ctx context.Context, email, username string) (bool, error)

	// Операции с баном
	Ban(ctx context.Context, userID string, banInfo *BanInfo) error
	Unban(ctx context.Context, userID string) error

	// Операции с подписками
	UpdateSubscription(ctx context.Context, userID string, subscription *SubscriptionInfo) error
}

func GenerateUUID() string {
	return uuid.New().String()
}

// AuditRepository определяет интерфейс для аудита и метаданных
type AuditRepository interface {
	// Операции с метаданными
	SaveMetadata(ctx context.Context, userID string, metadata map[string]string) error
	GetMetadata(ctx context.Context, userID string) (map[string]string, error)
	UpdateMetadata(ctx context.Context, userID string, metadata map[string]string) error
	DeleteMetadata(ctx context.Context, userID string, keys []string) error

	// Логирование действий
	LogActivity(ctx context.Context, activity *UserActivity) error
	GetUserActivities(ctx context.Context, userID string, limit int) ([]*UserActivity, error)

	// История подписок
	LogSubscriptionChange(ctx context.Context, entry *SubscriptionHistoryEntry) error
	GetSubscriptionHistory(ctx context.Context, userID string) ([]*SubscriptionHistoryEntry, error)

	// История банов
	LogBanChange(ctx context.Context, userID string, action string, details map[string]interface{}) error
	GetBanHistory(ctx context.Context, userID string) ([]map[string]interface{}, error)
}

// UserFilter фильтр для поиска пользователей
type UserFilter struct {
	Page      int
	PageSize  int
	Search    string
	Status    UserStatus
	Role      UserRole
	IsBanned  *bool
	SubStatus *SubscriptionStatus
	SubLevel  *SubscriptionLevel
}

// UserService определяет бизнес-логику работы с пользователями
type UserService interface {
	// CRUD операции
	Register(user *User) (*User, error)
	GetUser(id string) (*User, error)
	UpdateUser(user *User) (*User, error)
	DeleteUser(id string) error
	ListUsers(filter *UserFilter) ([]*User, int64, error)

	// Аутентификация и авторизация
	Authenticate(email, password string) (*User, string, error) // Возвращает пользователя и JWT токен
	ValidateToken(token string) (*User, error)
	ChangePassword(userID, currentPassword, newPassword string) error
	ResetPassword(email string) error

	// Бан-система
	BanUser(userID, reason, bannedBy string, duration *time.Duration) (*User, error)
	UnbanUser(userID, unbannedBy string) (*User, error)

	// Подписки
	UpdateSubscription(userID string, subscription *SubscriptionInfo) (*User, error)
	CancelSubscription(userID, reason string, immediate bool) (*User, error)
	CheckSubscriptionAccess(userID string, requiredLevel SubscriptionLevel, feature string) (bool, error)

	// Валидация
	ValidateEmail(email string) error
	ValidatePassword(password string) error

	// Health check
	HealthCheck() (bool, error)
}

// ===== Утилитарные функции =====

// NewUser создает нового пользователя с дефолтными значениями
func NewUser(email, password, name string) *User {
	now := time.Now()

	return &User{
		ID:           generateUUID(),
		Email:        email,
		Password:     password, // Ожидается уже хешированный пароль
		Name:         name,
		ServiceEmail: email,
		Status:       UserStatusActive,
		Role:         UserRoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		Metadata:     make(map[string]string),
	}
}

// NewBanInfo создает новую информацию о бане
func NewBanInfo(reason, bannedBy string, duration *time.Duration) *BanInfo {
	banInfo := &BanInfo{
		IsBanned: true,
		BannedAt: time.Now(),
		Reason:   reason,
		BannedBy: bannedBy,
	}

	if duration != nil {
		until := time.Now().Add(*duration)
		banInfo.BannedUntil = &until
	}

	return banInfo
}

// NewSubscriptionInfo создает новую информацию о подписке
func NewSubscriptionInfo(level SubscriptionLevel, trialDays int) *SubscriptionInfo {
	now := time.Now()
	var trialEnd *time.Time

	if trialDays > 0 {
		end := now.AddDate(0, 0, trialDays)
		trialEnd = &end
	}

	return &SubscriptionInfo{
		Status:            SubscriptionStatusTrial,
		Level:             level,
		SubscriptionStart: now,
		TrialEnd:          trialEnd,
		AutoRenew:         true,
		Features:          getDefaultFeaturesForLevel(level),
	}
}

// generateUUID генерирует UUID для пользователя
func generateUUID() string {
	// В реальном проекте используйте github.com/google/uuid
	return "generated-uuid-" + time.Now().Format("20060102150405")
}

// getDefaultFeaturesForLevel возвращает фичи по умолчанию для уровня подписки
func getDefaultFeaturesForLevel(level SubscriptionLevel) []string {
	switch level {
	case SubscriptionLevelFree:
		return []string{"basic_access", "read_only"}
	case SubscriptionLevelBasic:
		return []string{"basic_access", "create_content", "basic_analytics"}
	case SubscriptionLevelStandard:
		return []string{"basic_access", "create_content", "advanced_analytics", "export_data"}
	case SubscriptionLevelPro:
		return []string{"basic_access", "create_content", "advanced_analytics", "export_data", "api_access", "priority_support"}
	case SubscriptionLevelPremium:
		return []string{"all_features", "dedicated_support", "custom_integrations"}
	default:
		return []string{"basic_access"}
	}
}
