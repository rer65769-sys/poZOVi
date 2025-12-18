package server

import (
	"context"
	"errors"
	"fmt"
	"time"
	"userservice/internal/config"
	"userservice/internal/domain"
	"userservice/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   domain.UserRepository
	auditRepo  domain.AuditRepository
	jwtManager *jwt.JWTManager
	config     *config.Config
}

// HealthCheck implements [domain.UserService].
func (s *UserService) HealthCheck() (bool, error) {
	panic("unimplemented")
}

func NewUserService(userRepo domain.UserRepository, auditRepo domain.AuditRepository, jwtManager *jwt.JWTManager, cfg *config.Config) *UserService {
	return &UserService{
		userRepo:   userRepo,
		auditRepo:  auditRepo,
		jwtManager: jwtManager,
		config:     cfg,
	}
}

func (s *UserService) Register(user *domain.User) (*domain.User, error) {
	ctx := context.Background()

	// Проверяем, существует ли пользователь
	exists, err := s.userRepo.Exists(ctx, user.Email, user.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrUserAlreadyExists
	}

	// Хешируем пароль (если он еще не хешированный)
	// Проверяем, не хешированный ли уже пароль
	if !isHashedPassword(user.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	// Устанавливаем значения по умолчанию
	user.ID = domain.GenerateUUID()
	user.Status = domain.UserStatusActive
	user.Role = domain.UserRoleUser
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Создаем пользователя
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Сохраняем метаданные в MongoDB
	if len(user.Metadata) > 0 {
		if err := s.auditRepo.SaveMetadata(ctx, user.ID, user.Metadata); err != nil {
			// Логируем ошибку, но не прерываем регистрацию
			fmt.Printf("Warning: failed to save metadata: %v\n", err)
		}
	}

	// Логируем активность
	activity := domain.NewUserActivity(user.ID, domain.ActivityTypeLogin, "", "", "")
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	// Не возвращаем пароль
	user.Password = ""
	return user, nil
}

func (s *UserService) GetUser(id string) (*domain.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверяем, не забанен ли пользователь
	if user.IsBanned() {
		return nil, domain.ErrUserBanned
	}

	// Получаем метаданные из MongoDB
	metadata, err := s.auditRepo.GetMetadata(ctx, id)
	if err == nil && len(metadata) > 0 {
		user.Metadata = metadata
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateUser(user *domain.User) (*domain.User, error) {
	ctx := context.Background()

	// Получаем существующего пользователя
	existingUser, err := s.userRepo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Обновляем только разрешенные поля
	existingUser.Name = user.Name
	existingUser.Email = user.Email
	existingUser.Phone = user.Phone
	existingUser.ServiceEmail = user.ServiceEmail
	existingUser.UpdatedAt = time.Now()

	// Если передан пароль, хешируем его
	if user.Password != "" && !isHashedPassword(user.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		existingUser.Password = string(hashedPassword)
	}

	if err := s.userRepo.Update(ctx, existingUser); err != nil {
		return nil, err
	}

	// Обновляем метаданные в MongoDB
	if len(user.Metadata) > 0 {
		if err := s.auditRepo.UpdateMetadata(ctx, user.ID, user.Metadata); err != nil {
			fmt.Printf("Warning: failed to update metadata: %v\n", err)
		}
	}

	// Логируем активность
	activity := domain.NewUserActivity(user.ID, domain.ActivityTypeProfileUpdate, "", "", "")
	activity.AddDetail("fields_updated", "name, email, phone")
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	existingUser.Password = ""
	return existingUser, nil
}

func (s *UserService) DeleteUser(id string) error {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Мягкое удаление
	user.Status = domain.UserStatusDeleted
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) ListUsers(filter *domain.UserFilter) ([]*domain.User, int64, error) {
	ctx := context.Background()

	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Убираем пароли
	for _, user := range users {
		user.Password = ""
	}

	return users, total, nil
}

func (s *UserService) Authenticate(email, password string) (*domain.User, string, error) {
	ctx := context.Background()

	// Находим пользователя
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Проверяем пароль
	if !isHashedPassword(password) {
		// Пароль не хешированный, сравниваем с хешем
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			return nil, "", domain.ErrInvalidCredentials
		}
	} else {
		// Пароль уже хешированный, просто сравниваем
		if user.Password != password {
			return nil, "", domain.ErrInvalidCredentials
		}
	}

	// Проверяем, не забанен ли пользователь
	if user.IsBanned() {
		return nil, "", domain.ErrUserBanned
	}

	// Генерируем JWT токен
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	// Обновляем время последнего входа
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		fmt.Printf("Warning: failed to update last login: %v\n", err)
	}

	// Логируем активность
	activity := domain.NewUserActivity(user.ID, domain.ActivityTypeLogin, "", "", "")
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	// Убираем пароль из ответа
	user.Password = ""
	return user, token, nil
}

func (s *UserService) ValidateToken(token string) (*domain.User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.GetUser(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ChangePassword(userID, currentPassword, newPassword string) error {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Проверяем текущий пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Логируем активность
	activity := domain.NewUserActivity(userID, domain.ActivityTypePasswordChange, "", "", "")
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	return nil
}

func (s *UserService) ResetPassword(email string) error {
	// В реальном проекте здесь была бы логика отправки email с ссылкой для сброса пароля
	// Для примера просто возвращаем nil
	return nil
}

func (s *UserService) BanUser(userID, reason, bannedBy string, duration *time.Duration) (*domain.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь не является администратором
	if user.IsAdmin() {
		return nil, errors.New("cannot ban admin user")
	}

	// Создаем информацию о бане
	banInfo := domain.NewBanInfo(reason, bannedBy, duration)
	user.BanInfo = banInfo

	// Устанавливаем соответствующий статус
	if duration != nil {
		user.Status = domain.UserStatusBannedTemporarily
	} else {
		user.Status = domain.UserStatusBannedPermanently
	}

	if err := s.userRepo.Ban(ctx, userID, banInfo); err != nil {
		return nil, err
	}

	// Логируем в MongoDB
	details := map[string]interface{}{
		"reason":    reason,
		"banned_by": bannedBy,
		"duration":  duration,
		"user_id":   userID,
	}
	if err := s.auditRepo.LogBanChange(ctx, userID, "ban", details); err != nil {
		fmt.Printf("Warning: failed to log ban change: %v\n", err)
	}

	// Логируем активность
	activity := domain.NewUserActivity(userID, domain.ActivityTypeBan, "", "", "")
	activity.AddDetail("reason", reason)
	activity.AddDetail("banned_by", bannedBy)
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) UnbanUser(userID, unbannedBy string) (*domain.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.BanInfo == nil || !user.BanInfo.IsBanned {
		return nil, errors.New("user is not banned")
	}

	// Разбаниваем пользователя
	user.BanInfo.Unban()
	user.Status = domain.UserStatusActive

	if err := s.userRepo.Unban(ctx, userID); err != nil {
		return nil, err
	}

	// Логируем в MongoDB
	details := map[string]interface{}{
		"unbanned_by": unbannedBy,
		"user_id":     userID,
	}
	if err := s.auditRepo.LogBanChange(ctx, userID, "unban", details); err != nil {
		fmt.Printf("Warning: failed to log ban change: %v\n", err)
	}

	// Логируем активность
	activity := domain.NewUserActivity(userID, domain.ActivityTypeUnban, "", "", "")
	activity.AddDetail("unbanned_by", unbannedBy)
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) UpdateSubscription(userID string, subscription *domain.SubscriptionInfo) (*domain.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Сохраняем старые значения для истории
	var oldLevel domain.SubscriptionLevel
	var oldStatus domain.SubscriptionStatus
	if user.Subscription != nil {
		oldLevel = user.Subscription.Level
		oldStatus = user.Subscription.Status
	}

	user.Subscription = subscription

	if err := s.userRepo.UpdateSubscription(ctx, userID, subscription); err != nil {
		return nil, err
	}

	// Логируем изменение подписки
	entry := domain.NewSubscriptionHistoryEntry(
		userID,
		oldLevel,
		subscription.Level,
		oldStatus,
		subscription.Status,
		"Subscription updated",
		"system",
	)
	if err := s.auditRepo.LogSubscriptionChange(ctx, entry); err != nil {
		fmt.Printf("Warning: failed to log subscription change: %v\n", err)
	}

	// Логируем активность
	activity := domain.NewUserActivity(userID, domain.ActivityTypeSubscriptionStart, "", "", "")
	activity.AddDetail("level", string(subscription.Level))
	activity.AddDetail("status", string(subscription.Status))
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) CancelSubscription(userID, reason string, immediate bool) (*domain.User, error) {
	ctx := context.Background()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.Subscription == nil {
		return nil, errors.New("user does not have an active subscription")
	}

	// Сохраняем старые значения для истории
	oldLevel := user.Subscription.Level
	oldStatus := user.Subscription.Status

	// Отменяем подписку
	user.Subscription.Cancel(reason, immediate)

	if err := s.userRepo.UpdateSubscription(ctx, userID, user.Subscription); err != nil {
		return nil, err
	}

	// Логируем изменение подписки
	entry := domain.NewSubscriptionHistoryEntry(
		userID,
		oldLevel,
		user.Subscription.Level,
		oldStatus,
		user.Subscription.Status,
		reason,
		"system",
	)
	if err := s.auditRepo.LogSubscriptionChange(ctx, entry); err != nil {
		fmt.Printf("Warning: failed to log subscription change: %v\n", err)
	}

	// Логируем активность
	activity := domain.NewUserActivity(userID, domain.ActivityTypeSubscriptionEnd, "", "", "")
	activity.AddDetail("reason", reason)
	activity.AddDetail("immediate", immediate)
	if err := s.auditRepo.LogActivity(ctx, activity); err != nil {
		fmt.Printf("Warning: failed to log activity: %v\n", err)
	}

	user.Password = ""
	return user, nil
}

func (s *UserService) CheckSubscriptionAccess(userID string, requiredLevel domain.SubscriptionLevel, feature string) (bool, error) {
	user, err := s.GetUser(userID)
	if err != nil {
		return false, err
	}

	if user.Subscription == nil {
		return false, domain.ErrSubscriptionRequired
	}

	// Проверяем уровень подписки
	if user.Subscription.Level < requiredLevel {
		return false, domain.ErrSubscriptionRequired
	}

	// Проверяем доступ к конкретной фиче
	if feature != "" && !user.CanAccessFeature(feature) {
		return false, domain.ErrSubscriptionRequired
	}

	return true, nil
}

func (s *UserService) ValidateEmail(email string) error {
	// Простая проверка email
	if len(email) < 3 || len(email) > 255 {
		return errors.New("email must be between 3 and 255 characters")
	}
	// В реальном проекте добавьте более сложную проверку
	return nil
}

func (s *UserService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// В реальном проекте добавьте более сложные правила
	return nil
}

// func (s *UserService) HealthCheck() (bool, error) {
// 	ctx := context.Background()

// 	// Проверяем соединение с PostgreSQL
// 	if err := s.userRepo.Ping(ctx); err != nil {
// 		return false, fmt.Errorf("postgres connection failed: %v", err)
// 	}

// 	// Проверяем соединение с MongoDB
// 	if err := s.auditRepo.Ping(ctx); err != nil {
// 		return false, fmt.Errorf("mongodb connection failed: %v", err)
// 	}

// 	return true, nil
// }

// isHashedPassword проверяет, является ли пароль уже хешированным
func isHashedPassword(password string) bool {
	// BCrypt хеши начинаются с $2a$, $2b$, $2x$ или $2y$
	if len(password) == 60 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2x$" || password[:4] == "$2y$") {
		return true
	}
	return false
}
