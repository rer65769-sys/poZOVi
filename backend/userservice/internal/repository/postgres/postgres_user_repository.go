package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"userservice/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresUserRepository - репозиторий для PostgreSQL
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository создает новый репозиторий PostgreSQL
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// UserDBModel - модель пользователя в базе данных
type UserDBModel struct {
	ID                 string         `db:"id"`
	ServiceEmail       string         `db:"service_email"`
	Name               string         `db:"name"`
	Password           string         `db:"password"`
	Email              string         `db:"email"`
	Phone              string         `db:"phone"`
	Status             string         `db:"status"`
	Role               string         `db:"role"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	LastLoginAt        sql.NullTime   `db:"last_login_at"`
	BanInfo            sql.NullString `db:"ban_info"`     // JSON в базе
	Subscription       sql.NullString `db:"subscription"` // JSON в базе
	IsBanned           bool           `db:"is_banned"`
	SubscriptionStatus string         `db:"subscription_status"`
	SubscriptionLevel  string         `db:"subscription_level"`
	SubscriptionEnd    sql.NullTime   `db:"subscription_end"`
}

// ToDomain преобразует DB модель в доменную
func (dbUser *UserDBModel) ToDomain() (*domain.User, error) {
	user := &domain.User{
		ID:           dbUser.ID,
		ServiceEmail: dbUser.ServiceEmail,
		Name:         dbUser.Name,
		Password:     dbUser.Password,
		Email:        dbUser.Email,
		Phone:        dbUser.Phone,
		Status:       domain.UserStatus(dbUser.Status),
		Role:         domain.UserRole(dbUser.Role),
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}

	if dbUser.LastLoginAt.Valid {
		user.LastLoginAt = &dbUser.LastLoginAt.Time
	}

	// Парсим BanInfo из JSON
	if dbUser.BanInfo.Valid && dbUser.BanInfo.String != "" {
		var banInfo domain.BanInfo
		if err := json.Unmarshal([]byte(dbUser.BanInfo.String), &banInfo); err == nil {
			user.BanInfo = &banInfo
		}
	}

	// Парсим Subscription из JSON
	if dbUser.Subscription.Valid && dbUser.Subscription.String != "" {
		var subscription domain.SubscriptionInfo
		if err := json.Unmarshal([]byte(dbUser.Subscription.String), &subscription); err == nil {
			user.Subscription = &subscription
		}
	}

	return user, nil
}

// FromDomain преобразует доменную модель в DB модель
func (dbUser *UserDBModel) FromDomain(user *domain.User) error {
	dbUser.ID = user.ID
	dbUser.ServiceEmail = user.ServiceEmail
	dbUser.Name = user.Name
	dbUser.Password = user.Password
	dbUser.Email = user.Email
	dbUser.Phone = user.Phone
	dbUser.Status = string(user.Status)
	dbUser.Role = string(user.Role)
	dbUser.CreatedAt = user.CreatedAt
	dbUser.UpdatedAt = user.UpdatedAt
	dbUser.IsBanned = user.BanInfo != nil && user.BanInfo.IsBanned

	if user.LastLoginAt != nil {
		dbUser.LastLoginAt = sql.NullTime{Time: *user.LastLoginAt, Valid: true}
	}

	// Сериализуем BanInfo в JSON
	if user.BanInfo != nil {
		banInfoJSON, err := json.Marshal(user.BanInfo)
		if err == nil {
			dbUser.BanInfo = sql.NullString{String: string(banInfoJSON), Valid: true}
		}
	}

	// Сериализуем Subscription в JSON
	if user.Subscription != nil {
		subscriptionJSON, err := json.Marshal(user.Subscription)
		if err == nil {
			dbUser.Subscription = sql.NullString{String: string(subscriptionJSON), Valid: true}
			dbUser.SubscriptionStatus = string(user.Subscription.Status)
			dbUser.SubscriptionLevel = string(user.Subscription.Level)

			if user.Subscription.SubscriptionEnd != nil {
				dbUser.SubscriptionEnd = sql.NullTime{
					Time:  *user.Subscription.SubscriptionEnd,
					Valid: true,
				}
			}
		}
	}

	return nil
}

// Create создает нового пользователя
func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	dbUser := &UserDBModel{}
	if err := dbUser.FromDomain(user); err != nil {
		return fmt.Errorf("failed to convert domain to db model: %w", err)
	}

	query := `
		INSERT INTO users (
			id, service_email, name, password, email, phone, status, role,
			created_at, updated_at, last_login_at, ban_info, subscription,
			is_banned, subscription_status, subscription_level, subscription_end
		) VALUES (
			:id, :service_email, :name, :password, :email, :phone, :status, :role,
			:created_at, :updated_at, :last_login_at, :ban_info, :subscription,
			:is_banned, :subscription_status, :subscription_level, :subscription_end
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, dbUser)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.NewUserAlreadyExistsError(user.Email, user.Name)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// FindByID находит пользователя по ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var dbUser UserDBModel

	query := `SELECT * FROM users WHERE id = $1 AND status != $2`
	err := r.db.GetContext(ctx, &dbUser, query, id, domain.UserStatusDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewUserNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	user, err := dbUser.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert db model to domain: %w", err)
	}

	return user, nil
}

// FindByEmail находит пользователя по email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dbUser UserDBModel

	query := `SELECT * FROM users WHERE email = $1 AND status != $2`
	err := r.db.GetContext(ctx, &dbUser, query, email, domain.UserStatusDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewUserNotFoundByEmailError(email)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user, err := dbUser.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert db model to domain: %w", err)
	}

	return user, nil
}

// Update обновляет пользователя
func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	dbUser := &UserDBModel{}
	if err := dbUser.FromDomain(user); err != nil {
		return fmt.Errorf("failed to convert domain to db model: %w", err)
	}

	query := `
		UPDATE users SET
			service_email = :service_email,
			name = :name,
			password = :password,
			email = :email,
			phone = :phone,
			status = :status,
			role = :role,
			updated_at = :updated_at,
			last_login_at = :last_login_at,
			ban_info = :ban_info,
			subscription = :subscription,
			is_banned = :is_banned,
			subscription_status = :subscription_status,
			subscription_level = :subscription_level,
			subscription_end = :subscription_end
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, dbUser)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.NewUserNotFoundError(user.ID)
	}

	return nil
}

// Delete удаляет пользователя (мягкое удаление)
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users SET 
			status = $1, 
			updated_at = $2,
			deleted_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		domain.UserStatusDeleted,
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return domain.NewUserNotFoundError(id)
	}

	return nil
}

// List возвращает список пользователей с фильтрацией и пагинацией
func (r *PostgresUserRepository) List(ctx context.Context, filter *domain.UserFilter) ([]*domain.User, int64, error) {
	// Строим WHERE условия
	var conditions []string
	var args []interface{}
	argPos := 1

	conditions = append(conditions, "status != $1")
	args = append(args, domain.UserStatusDeleted)
	argPos++

	if filter.Search != "" {
		conditions = append(conditions,
			fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)",
				argPos, argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	if filter.Status != "" && filter.Status != domain.UserStatusUnspecified {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPos))
		args = append(args, string(filter.Status))
		argPos++
	}

	if filter.Role != "" && filter.Role != domain.UserRoleUnspecified {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argPos))
		args = append(args, string(filter.Role))
		argPos++
	}

	if filter.IsBanned != nil {
		conditions = append(conditions, fmt.Sprintf("is_banned = $%d", argPos))
		args = append(args, *filter.IsBanned)
		argPos++
	}

	if filter.SubStatus != nil && *filter.SubStatus != "" {
		conditions = append(conditions, fmt.Sprintf("subscription_status = $%d", argPos))
		args = append(args, string(*filter.SubStatus))
		argPos++
	}

	if filter.SubLevel != nil && *filter.SubLevel != "" {
		conditions = append(conditions, fmt.Sprintf("subscription_level = $%d", argPos))
		args = append(args, string(*filter.SubLevel))
		argPos++
	}

	// Подсчет общего количества
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", strings.Join(conditions, " AND "))
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Получение данных с пагинацией
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	query := fmt.Sprintf(`
		SELECT * FROM users 
		WHERE %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`,
		strings.Join(conditions, " AND "), argPos, argPos+1)

	var dbUsers []UserDBModel
	err = r.db.SelectContext(ctx, &dbUsers, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Конвертация в доменные модели
	users := make([]*domain.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		user, err := dbUser.ToDomain()
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, total, nil
}

// FindByPhone находит пользователя по телефону
func (r *PostgresUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var dbUser UserDBModel

	query := `SELECT * FROM users WHERE phone = $1 AND status != $2`
	err := r.db.GetContext(ctx, &dbUser, query, phone, domain.UserStatusDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewUserNotFoundError("phone: " + phone)
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	user, err := dbUser.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert db model to domain: %w", err)
	}

	return user, nil
}

// Exists проверяет существование пользователя по email или username
func (r *PostgresUserRepository) Exists(ctx context.Context, email, username string) (bool, error) {
	var count int

	query := `SELECT COUNT(*) FROM users WHERE (email = $1 OR name = $2) AND status != $3`
	err := r.db.GetContext(ctx, &count, query, email, username, domain.UserStatusDeleted)

	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

// Ban банит пользователя
func (r *PostgresUserRepository) Ban(ctx context.Context, userID string, banInfo *domain.BanInfo) error {
	banInfoJSON, err := json.Marshal(banInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal ban info: %w", err)
	}

	query := `
		UPDATE users SET 
			ban_info = $1,
			is_banned = true,
			status = CASE 
				WHEN $2::jsonb->>'banned_until' IS NOT NULL THEN $3
				ELSE $4
			END,
			updated_at = $5
		WHERE id = $6
	`

	_, err = r.db.ExecContext(ctx, query,
		string(banInfoJSON),
		string(banInfoJSON),
		domain.UserStatusBannedTemporarily,
		domain.UserStatusBannedPermanently,
		time.Now(),
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	return nil
}

// Unban разбанивает пользователя
func (r *PostgresUserRepository) Unban(ctx context.Context, userID string) error {
	query := `
		UPDATE users SET 
			ban_info = NULL,
			is_banned = false,
			status = $1,
			updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query,
		domain.UserStatusActive,
		time.Now(),
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to unban user: %w", err)
	}

	return nil
}

// UpdateSubscription обновляет подписку пользователя
func (r *PostgresUserRepository) UpdateSubscription(ctx context.Context, userID string, subscription *domain.SubscriptionInfo) error {
	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	query := `
		UPDATE users SET 
			subscription = $1,
			subscription_status = $2,
			subscription_level = $3,
			subscription_end = $4,
			updated_at = $5
		WHERE id = $6
	`

	var subscriptionEnd interface{}
	if subscription.SubscriptionEnd != nil {
		subscriptionEnd = subscription.SubscriptionEnd
	} else {
		subscriptionEnd = nil
	}

	_, err = r.db.ExecContext(ctx, query,
		string(subscriptionJSON),
		string(subscription.Status),
		string(subscription.Level),
		subscriptionEnd,
		time.Now(),
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// CancelSubscription отменяет подписку
func (r *PostgresUserRepository) CancelSubscription(ctx context.Context, userID string, reason string, immediate bool) error {
	// Получаем текущую подписку
	user, err := r.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Subscription == nil {
		// Используйте новую ошибку вместо NewDomainError
		return domain.NewSubscriptionNotFoundError(userID)
	}

	// Обновляем подписку
	user.Subscription.Cancel(reason, immediate)

	// Сохраняем изменения
	return r.UpdateSubscription(ctx, userID, user.Subscription)
}

// UpdateLastLogin обновляет время последнего входа
func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users SET 
			last_login_at = $1,
			updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// Ping проверяет соединение с базой данных
func (r *PostgresUserRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
