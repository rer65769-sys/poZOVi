package domain

import (
	"errors"
	"fmt"
)

// ===== Базовые ошибки =====

// DomainError - базовая ошибка доменного слоя
type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (код: %s)", e.Message, e.Cause.Error(), e.Code)
	}
	return fmt.Sprintf("%s (код: %s)", e.Message, e.Code)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError создает новую доменную ошибку
func NewDomainError(code, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ===== Конкретные ошибки пользователя =====

// UserError коды ошибок пользователей
const (
	ErrCodeUserNotFound         = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists    = "USER_ALREADY_EXISTS"
	ErrCodeInvalidCredentials   = "INVALID_CREDENTIALS"
	ErrCodeUserBanned           = "USER_BANNED"
	ErrCodeUserInactive         = "USER_INACTIVE"
	ErrCodeInvalidEmail         = "INVALID_EMAIL"
	ErrCodeInvalidPassword      = "INVALID_PASSWORD"
	ErrCodeEmailNotVerified     = "EMAIL_NOT_VERIFIED"
	ErrCodePhoneNotVerified     = "PHONE_NOT_VERIFIED"
	ErrCodePermissionDenied     = "PERMISSION_DENIED"
	ErrCodeInvalidRole          = "INVALID_ROLE"
	ErrCodeInvalidStatus        = "INVALID_STATUS"
	ErrCodeSubscriptionNotFound = "SUBSCRIPTION_NOT_FOUND"
)

// Обертки для стандартных ошибок
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserBanned           = errors.New("user is banned")
	ErrUserInactive         = errors.New("user is inactive")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrPhoneNotVerified     = errors.New("phone not verified")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// Функции для создания ошибок с дополнительным контекстом
func NewUserNotFoundError(userID string) *DomainError {
	return NewDomainError(
		ErrCodeUserNotFound,
		fmt.Sprintf("Пользователь с ID '%s' не найден", userID),
		nil,
	)
}

func NewUserNotFoundByEmailError(email string) *DomainError {
	return NewDomainError(
		ErrCodeUserNotFound,
		fmt.Sprintf("Пользователь с email '%s' не найден", email),
		nil,
	)
}

func NewUserNotFoundErrorByPhone(phone string) *DomainError {
	return NewDomainError(
		ErrCodeUserNotFound,
		fmt.Sprintf("Пользователь с телефоном '%s' не найден", phone),
		nil,
	)
}

func NewUserAlreadyExistsError(email, username string) *DomainError {
	msg := "Пользователь уже существует"
	if email != "" && username != "" {
		msg = fmt.Sprintf("Пользователь с email '%s' или username '%s' уже существует", email, username)
	} else if email != "" {
		msg = fmt.Sprintf("Пользователь с email '%s' уже существует", email)
	} else if username != "" {
		msg = fmt.Sprintf("Пользователь с username '%s' уже существует", username)
	}

	return NewDomainError(ErrCodeUserAlreadyExists, msg, nil)
}

func NewInvalidCredentialsError(cause error) *DomainError {
	return NewDomainError(ErrCodeInvalidCredentials, "Неверные учетные данные", cause)
}

func NewUserBannedError(banInfo *BanInfo) *DomainError {
	msg := "Пользователь забанен"
	if banInfo != nil && banInfo.Reason != "" {
		msg = fmt.Sprintf("Пользователь забанен. Причина: %s", banInfo.Reason)
		if banInfo.BannedUntil != nil {
			msg += fmt.Sprintf(". Бан действует до: %s", banInfo.BannedUntil.Format("2006-01-02 15:04:05"))
		} else {
			msg += " (перманентный бан)"
		}
	}

	return NewDomainError(ErrCodeUserBanned, msg, nil)
}

func NewPermissionDeniedError(requiredRole UserRole, actualRole UserRole) *DomainError {
	msg := "Доступ запрещен"
	if requiredRole != "" {
		msg = fmt.Sprintf("Требуется роль: %s. Ваша роль: %s", requiredRole, actualRole)
	}

	return NewDomainError(ErrCodePermissionDenied, msg, nil)
}

// NewSubscriptionNotFoundError создает ошибку ненайденной подписки
func NewSubscriptionNotFoundError(userID string) *DomainError {
	return NewDomainError(
		ErrCodeSubscriptionNotFound,
		fmt.Sprintf("Подписка для пользователя с ID '%s' не найдена", userID),
		nil,
	)
}

// ===== Ошибки банов =====

// BanError коды ошибок банов
const (
	ErrCodeBanAlreadyActive   = "BAN_ALREADY_ACTIVE"
	ErrCodeBanNotFound        = "BAN_NOT_FOUND"
	ErrCodeInvalidBanDuration = "INVALID_BAN_DURATION"
	ErrCodeInvalidBanReason   = "INVALID_BAN_REASON"
	ErrCodeSelfBanNotAllowed  = "SELF_BAN_NOT_ALLOWED"
	ErrCodeAdminBanNotAllowed = "ADMIN_BAN_NOT_ALLOWED"
)

// Обертки для ошибок банов
var (
	ErrBanAlreadyActive   = NewDomainError(ErrCodeBanAlreadyActive, "Пользователь уже забанен", nil)
	ErrBanNotFound        = NewDomainError(ErrCodeBanNotFound, "Бан не найден", nil)
	ErrInvalidBanDuration = NewDomainError(ErrCodeInvalidBanDuration, "Некорректная длительность бана", nil)
	ErrInvalidBanReason   = NewDomainError(ErrCodeInvalidBanReason, "Некорректная причина бана", nil)
	ErrSelfBanNotAllowed  = NewDomainError(ErrCodeSelfBanNotAllowed, "Нельзя забанить самого себя", nil)
	ErrAdminBanNotAllowed = NewDomainError(ErrCodeAdminBanNotAllowed, "Нельзя забанить администратора", nil)
)

// Функции для создания ошибок банов с контекстом
func NewBanAlreadyActiveError(userID string) *DomainError {
	return NewDomainError(
		ErrCodeBanAlreadyActive,
		fmt.Sprintf("Пользователь '%s' уже забанен", userID),
		nil,
	)
}

func NewAdminBanNotAllowedError(adminRole UserRole) *DomainError {
	return NewDomainError(
		ErrCodeAdminBanNotAllowed,
		fmt.Sprintf("Нельзя забанить пользователя с ролью администратора (%s)", adminRole),
		nil,
	)
}

// ===== Ошибки подписок =====

// SubscriptionError коды ошибок подписок
const (
	ErrCodeSubscriptionRequired      = "SUBSCRIPTION_REQUIRED"
	ErrCodeInvalidSubscriptionLevel  = "INVALID_SUBSCRIPTION_LEVEL"
	ErrCodeSubscriptionExpired       = "SUBSCRIPTION_EXPIRED"
	ErrCodeTrialExpired              = "TRIAL_EXPIRED"
	ErrCodePaymentRequired           = "PAYMENT_REQUIRED"
	ErrCodeInvalidPaymentMethod      = "INVALID_PAYMENT_METHOD"
	ErrCodeFeatureNotAvailable       = "FEATURE_NOT_AVAILABLE"
	ErrCodeSubscriptionAlreadyActive = "SUBSCRIPTION_ALREADY_ACTIVE"
	ErrCodeInvalidAmount             = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency           = "INVALID_CURRENCY"
)

// Обертки для ошибок подписок
var (
	ErrSubscriptionRequired      = NewDomainError(ErrCodeSubscriptionRequired, "Требуется подписка", nil)
	ErrInvalidSubscriptionLevel  = NewDomainError(ErrCodeInvalidSubscriptionLevel, "Некорректный уровень подписки", nil)
	ErrSubscriptionExpired       = NewDomainError(ErrCodeSubscriptionExpired, "Подписка истекла", nil)
	ErrTrialExpired              = NewDomainError(ErrCodeTrialExpired, "Пробный период истек", nil)
	ErrPaymentRequired           = NewDomainError(ErrCodePaymentRequired, "Требуется оплата", nil)
	ErrInvalidPaymentMethod      = NewDomainError(ErrCodeInvalidPaymentMethod, "Некорректный способ оплаты", nil)
	ErrFeatureNotAvailable       = NewDomainError(ErrCodeFeatureNotAvailable, "Функция недоступна для вашей подписки", nil)
	ErrSubscriptionAlreadyActive = NewDomainError(ErrCodeSubscriptionAlreadyActive, "Подписка уже активна", nil)
	ErrInvalidAmount             = NewDomainError(ErrCodeInvalidAmount, "Некорректная сумма", nil)
	ErrInvalidCurrency           = NewDomainError(ErrCodeInvalidCurrency, "Некорректная валюта", nil)
)

// Функции для создания ошибок подписок с контекстом
func NewSubscriptionRequiredError(requiredLevel SubscriptionLevel) *DomainError {
	msg := "Требуется подписка"
	if requiredLevel != "" {
		msg = fmt.Sprintf("Требуется подписка уровня '%s'", requiredLevel)
	}

	return NewDomainError(ErrCodeSubscriptionRequired, msg, nil)
}

func NewFeatureNotAvailableError(feature string, currentLevel SubscriptionLevel) *DomainError {
	return NewDomainError(
		ErrCodeFeatureNotAvailable,
		fmt.Sprintf("Функция '%s' недоступна для уровня подписки '%s'", feature, currentLevel),
		nil,
	)
}

func NewSubscriptionExpiredError(subscriptionEnd string) *DomainError {
	msg := "Подписка истекла"
	if subscriptionEnd != "" {
		msg = fmt.Sprintf("Подписка истекла %s", subscriptionEnd)
	}

	return NewDomainError(ErrCodeSubscriptionExpired, msg, nil)
}

// ===== Ошибки валидации =====

// ValidationError коды ошибок валидации
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeRequiredField    = "REQUIRED_FIELD"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeInvalidLength    = "INVALID_LENGTH"
	ErrCodeDuplicateValue   = "DUPLICATE_VALUE"
)

// ValidationError - ошибка валидации с деталями
type ValidationError struct {
	DomainError
	Field   string
	Details map[string]interface{}
}

func NewValidationError(field, message string, details map[string]interface{}) *ValidationError {
	return &ValidationError{
		DomainError: DomainError{
			Code:    ErrCodeValidationFailed,
			Message: message,
		},
		Field:   field,
		Details: details,
	}
}

// Функции для создания ошибок валидации
func NewRequiredFieldError(field string) *ValidationError {
	return NewValidationError(
		field,
		fmt.Sprintf("Поле '%s' обязательно для заполнения", field),
		map[string]interface{}{
			"field": field,
			"type":  "required",
		},
	)
}

func NewInvalidFormatError(field, format string) *ValidationError {
	return NewValidationError(
		field,
		fmt.Sprintf("Поле '%s' имеет некорректный формат. Ожидается: %s", field, format),
		map[string]interface{}{
			"field":  field,
			"format": format,
			"type":   "format",
		},
	)
}

func NewInvalidLengthError(field string, min, max, actual int) *ValidationError {
	msg := fmt.Sprintf("Поле '%s' имеет некорректную длину", field)
	if min > 0 && max > 0 {
		msg = fmt.Sprintf("Поле '%s' должно содержать от %d до %d символов. Текущая длина: %d", field, min, max, actual)
	} else if min > 0 {
		msg = fmt.Sprintf("Поле '%s' должно содержать минимум %d символов. Текущая длина: %d", field, min, actual)
	} else if max > 0 {
		msg = fmt.Sprintf("Поле '%s' должно содержать максимум %d символов. Текущая длина: %d", field, max, actual)
	}

	return NewValidationError(
		field,
		msg,
		map[string]interface{}{
			"field":  field,
			"min":    min,
			"max":    max,
			"actual": actual,
			"type":   "length",
		},
	)
}

func NewDuplicateValueError(field, value string) *ValidationError {
	return NewValidationError(
		field,
		fmt.Sprintf("Значение '%s' уже используется в поле '%s'", value, field),
		map[string]interface{}{
			"field": field,
			"value": value,
			"type":  "duplicate",
		},
	)
}

// ===== Ошибки токенов =====

// TokenError коды ошибок токенов
const (
	ErrCodeInvalidToken        = "INVALID_TOKEN"
	ErrCodeTokenExpired        = "TOKEN_EXPIRED"
	ErrCodeTokenRevoked        = "TOKEN_REVOKED"
	ErrCodeTokenNotProvided    = "TOKEN_NOT_PROVIDED"
	ErrCodeRefreshTokenInvalid = "REFRESH_TOKEN_INVALID"
	ErrCodeTokenGeneration     = "TOKEN_GENERATION_FAILED"
)

// Обертки для ошибок токенов
var (
	ErrInvalidToken        = NewDomainError(ErrCodeInvalidToken, "Некорректный токен", nil)
	ErrTokenExpired        = NewDomainError(ErrCodeTokenExpired, "Срок действия токена истек", nil)
	ErrTokenRevoked        = NewDomainError(ErrCodeTokenRevoked, "Токен отозван", nil)
	ErrTokenNotProvided    = NewDomainError(ErrCodeTokenNotProvided, "Токен не предоставлен", nil)
	ErrRefreshTokenInvalid = NewDomainError(ErrCodeRefreshTokenInvalid, "Некорректный refresh токен", nil)
	ErrTokenGeneration     = NewDomainError(ErrCodeTokenGeneration, "Ошибка генерации токена", nil)
)

// ===== Ошибки базы данных =====

// DatabaseError коды ошибок базы данных
const (
	ErrCodeDBConnection  = "DB_CONNECTION_FAILED"
	ErrCodeDBQuery       = "DB_QUERY_FAILED"
	ErrCodeDBTransaction = "DB_TRANSACTION_FAILED"
	ErrCodeDBConstraint  = "DB_CONSTRAINT_VIOLATION"
	ErrCodeDBDeadlock    = "DB_DEADLOCK"
	ErrCodeDBTimeout     = "DB_TIMEOUT"
)

// Функции для создания ошибок базы данных
func NewDBConnectionError(cause error) *DomainError {
	return NewDomainError(
		ErrCodeDBConnection,
		"Ошибка подключения к базе данных",
		cause,
	)
}

func NewDBQueryError(query string, cause error) *DomainError {
	return NewDomainError(
		ErrCodeDBQuery,
		fmt.Sprintf("Ошибка выполнения запроса: %s", query),
		cause,
	)
}

func NewDBConstraintError(constraint string, cause error) *DomainError {
	return NewDomainError(
		ErrCodeDBConstraint,
		fmt.Sprintf("Нарушение ограничения базы данных: %s", constraint),
		cause,
	)
}
