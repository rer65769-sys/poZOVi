package domain

import (
	users "userservice/gen/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto преобразует доменного User в protobuf User
func (u *User) ToProto() *users.User {
	protoUser := &users.User{
		Id:           u.ID,
		ServiceEmail: u.ServiceEmail,
		Name:         u.Name,
		Password:     u.Password,
		Email:        u.Email,
		Phone:        u.Phone,
		Status:       UserStatusToProto(u.Status),
		Role:         UserRoleToProto(u.Role),
		CreatedAt:    timestamppb.New(u.CreatedAt),
		UpdatedAt:    timestamppb.New(u.UpdatedAt),
		Metadata:     u.Metadata,
	}

	if u.LastLoginAt != nil {
		protoUser.LastLoginAt = timestamppb.New(*u.LastLoginAt)
	}

	if u.BanInfo != nil {
		protoUser.BanInfo = u.BanInfo.ToProto()
	}

	if u.Subscription != nil {
		protoUser.Subscription = u.Subscription.ToProto()
	}

	return protoUser
}

// ToProto преобразует BanInfo в protobuf BanInfo
func (b *BanInfo) ToProto() *users.BanInfo {
	protoBan := &users.BanInfo{
		IsBanned: b.IsBanned,
		BannedAt: timestamppb.New(b.BannedAt),
		Reason:   b.Reason,
		BannedBy: b.BannedBy,
	}

	if b.BannedUntil != nil {
		protoBan.BannedUntil = timestamppb.New(*b.BannedUntil)
	}

	return protoBan
}

// ToProto преобразует SubscriptionInfo в protobuf SubscriptionInfo
func (s *SubscriptionInfo) ToProto() *users.SubscriptionInfo {
	protoSub := &users.SubscriptionInfo{
		Status:            SubscriptionStatusToProto(s.Status),
		Level:             SubscriptionLevelToProto(s.Level),
		SubscriptionStart: timestamppb.New(s.SubscriptionStart),
		SubscriptionId:    s.SubscriptionID,
		PaymentMethod:     s.PaymentMethod,
		AutoRenew:         s.AutoRenew,
		Amount:            s.Amount,
		Currency:          s.Currency,
		CancelReason:      s.CancelReason,
		Features:          s.Features,
	}

	if s.SubscriptionEnd != nil {
		protoSub.SubscriptionEnd = timestamppb.New(*s.SubscriptionEnd)
	}

	if s.TrialEnd != nil {
		protoSub.TrialEnd = timestamppb.New(*s.TrialEnd)
	}

	if s.NextBillingDate != nil {
		protoSub.NextBillingDate = timestamppb.New(*s.NextBillingDate)
	}

	if s.CanceledAt != nil {
		protoSub.CanceledAt = timestamppb.New(*s.CanceledAt)
	}

	if s.GracePeriodEnd != nil {
		protoSub.GracePeriodEnd = timestamppb.New(*s.GracePeriodEnd)
	}

	return protoSub
}

// CreateUserRequestFromProto преобразует protobuf CreateUserRequest в доменную модель
func CreateUserRequestFromProto(req *users.CreateUserRequest) *User {
	user := &User{
		Name:         req.GetName(),
		Email:        req.GetEmail(),
		Password:     req.GetPassword(), // Уже хешированный
		Phone:        req.GetPhone(),
		ServiceEmail: req.GetServiceEmail(),
		Role:         UserRoleFromProto(req.GetRole()),
	}

	return user
}

// UpdateUserFromProto обновляет доменного User из protobuf UpdateUserRequest
func (u *User) UpdateFromProto(req *users.UpdateUserRequest) {
	if req.GetName() != "" {
		u.Name = req.GetName()
	}

	if req.GetEmail() != "" {
		u.Email = req.GetEmail()
	}

	if req.GetPassword() != "" {
		u.Password = req.GetPassword() // Уже хешированный
	}

	if req.GetPhone() != "" {
		u.Phone = req.GetPhone()
	}

	if req.GetServiceEmail() != "" {
		u.ServiceEmail = req.GetServiceEmail()
	}

	if req.GetStatus() != users.UserStatus_USER_STATUS_UNSPECIFIED {
		u.Status = UserStatusFromProto(req.GetStatus())
	}

	if req.GetRole() != users.UserRole_USER_ROLE_UNSPECIFIED {
		u.Role = UserRoleFromProto(req.GetRole())
	}

	// Обновляем метаданные
	if req.Metadata != nil {
		if u.Metadata == nil {
			u.Metadata = make(map[string]string)
		}
		for k, v := range req.Metadata {
			u.Metadata[k] = v
		}
	}
}

// ListUsersResponseToProto преобразует доменные данные в protobuf ListUsersResponse
func ListUsersResponseToProto(userses []*User, total int64, page, pageSize int32) *users.ListUsersResponse {
	var protoUsers []*users.User
	for _, user := range userses {
		protoUsers = append(protoUsers, user.ToProto())
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &users.ListUsersResponse{
		Users:      protoUsers,
		Total:      int32(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// UserStatusToProto преобразует доменный UserStatus в protobuf
func UserStatusToProto(status UserStatus) users.UserStatus {
	switch status {
	case UserStatusActive:
		return users.UserStatus_USER_STATUS_ACTIVE
	case UserStatusInactive:
		return users.UserStatus_USER_STATUS_INACTIVE
	case UserStatusSuspended:
		return users.UserStatus_USER_STATUS_SUSPENDED
	case UserStatusPending:
		return users.UserStatus_USER_STATUS_PENDING
	case UserStatusDeleted:
		return users.UserStatus_USER_STATUS_DELETED
	case UserStatusBannedPermanently:
		return users.UserStatus_USER_STATUS_BANNED_PERMANENTLY
	case UserStatusBannedTemporarily:
		return users.UserStatus_USER_STATUS_BANNED_TEMPORARILY
	case UserStatusBannedByAdmin:
		return users.UserStatus_USER_STATUS_BANNED_BY_ADMIN
	case UserStatusBannedBySystem:
		return users.UserStatus_USER_STATUS_BANNED_BY_SYSTEM
	case UserStatusBannedForSpam:
		return users.UserStatus_USER_STATUS_BANNED_FOR_SPAM
	case UserStatusBannedForAbuse:
		return users.UserStatus_USER_STATUS_BANNED_FOR_ABUSE
	case UserStatusBannedForFraud:
		return users.UserStatus_USER_STATUS_BANNED_FOR_FRAUD
	default:
		return users.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// UserStatusFromProto преобразует protobuf UserStatus в доменный
func UserStatusFromProto(protoStatus users.UserStatus) UserStatus {
	switch protoStatus {
	case users.UserStatus_USER_STATUS_ACTIVE:
		return UserStatusActive
	case users.UserStatus_USER_STATUS_INACTIVE:
		return UserStatusInactive
	case users.UserStatus_USER_STATUS_SUSPENDED:
		return UserStatusSuspended
	case users.UserStatus_USER_STATUS_PENDING:
		return UserStatusPending
	case users.UserStatus_USER_STATUS_DELETED:
		return UserStatusDeleted
	case users.UserStatus_USER_STATUS_BANNED_PERMANENTLY:
		return UserStatusBannedPermanently
	case users.UserStatus_USER_STATUS_BANNED_TEMPORARILY:
		return UserStatusBannedTemporarily
	case users.UserStatus_USER_STATUS_BANNED_BY_ADMIN:
		return UserStatusBannedByAdmin
	case users.UserStatus_USER_STATUS_BANNED_BY_SYSTEM:
		return UserStatusBannedBySystem
	case users.UserStatus_USER_STATUS_BANNED_FOR_SPAM:
		return UserStatusBannedForSpam
	case users.UserStatus_USER_STATUS_BANNED_FOR_ABUSE:
		return UserStatusBannedForAbuse
	case users.UserStatus_USER_STATUS_BANNED_FOR_FRAUD:
		return UserStatusBannedForFraud
	default:
		return UserStatusUnspecified
	}
}

// UserRoleToProto преобразует доменный UserRole в protobuf
func UserRoleToProto(role UserRole) users.UserRole {
	switch role {
	case UserRoleUser:
		return users.UserRole_USER_ROLE_USER
	case UserRoleModerator:
		return users.UserRole_USER_ROLE_MODERATOR
	case UserRoleAdmin:
		return users.UserRole_USER_ROLE_ADMIN
	case UserRoleSuperAdmin:
		return users.UserRole_USER_ROLE_SUPER_ADMIN
	case UserRoleBannedUser:
		return users.UserRole_USER_ROLE_BANNED_USER
	default:
		return users.UserRole_USER_ROLE_UNSPECIFIED
	}
}

// UserRoleFromProto преобразует protobuf UserRole в доменный
func UserRoleFromProto(protoRole users.UserRole) UserRole {
	switch protoRole {
	case users.UserRole_USER_ROLE_USER:
		return UserRoleUser
	case users.UserRole_USER_ROLE_MODERATOR:
		return UserRoleModerator
	case users.UserRole_USER_ROLE_ADMIN:
		return UserRoleAdmin
	case users.UserRole_USER_ROLE_SUPER_ADMIN:
		return UserRoleSuperAdmin
	case users.UserRole_USER_ROLE_BANNED_USER:
		return UserRoleBannedUser
	default:
		return UserRoleUnspecified
	}
}

// SubscriptionStatusToProto преобразует доменный SubscriptionStatus в protobuf
func SubscriptionStatusToProto(status SubscriptionStatus) users.SubscriptionStatus {
	switch status {
	case SubscriptionStatusInactive:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_INACTIVE
	case SubscriptionStatusTrial:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIAL
	case SubscriptionStatusActive:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE
	case SubscriptionStatusPastDue:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE
	case SubscriptionStatusCanceled:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED
	case SubscriptionStatusExpired:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_EXPIRED
	case SubscriptionStatusPaused:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_PAUSED
	case SubscriptionStatusPending:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_PENDING
	case SubscriptionStatusGracePeriod:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_GRACE_PERIOD
	case SubscriptionStatusUpgrading:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_UPGRADING
	case SubscriptionStatusDowngrading:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_DOWNGRADING
	default:
		return users.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED
	}
}

// SubscriptionStatusFromProto преобразует protobuf SubscriptionStatus в доменный
func SubscriptionStatusFromProto(protoStatus users.SubscriptionStatus) SubscriptionStatus {
	switch protoStatus {
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_INACTIVE:
		return SubscriptionStatusInactive
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIAL:
		return SubscriptionStatusTrial
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE:
		return SubscriptionStatusActive
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_PAST_DUE:
		return SubscriptionStatusPastDue
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED:
		return SubscriptionStatusCanceled
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_EXPIRED:
		return SubscriptionStatusExpired
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_PAUSED:
		return SubscriptionStatusPaused
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_PENDING:
		return SubscriptionStatusPending
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_GRACE_PERIOD:
		return SubscriptionStatusGracePeriod
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_UPGRADING:
		return SubscriptionStatusUpgrading
	case users.SubscriptionStatus_SUBSCRIPTION_STATUS_DOWNGRADING:
		return SubscriptionStatusDowngrading
	default:
		return SubscriptionStatusUnspecified
	}
}

// SubscriptionLevelToProto преобразует доменный SubscriptionLevel в protobuf
func SubscriptionLevelToProto(level SubscriptionLevel) users.SubscriptionLevel {
	switch level {
	case SubscriptionLevelFree:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_FREE
	case SubscriptionLevelBasic:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_BASIC
	case SubscriptionLevelStandard:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_STANDARD
	case SubscriptionLevelPro:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_PRO
	case SubscriptionLevelPremium:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_PREMIUM
	case SubscriptionLevelEnterprise:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_ENTERPRISE
	case SubscriptionLevelLifetime:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_LIFETIME
	case SubscriptionLevelStarter:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_STARTER
	case SubscriptionLevelBusiness:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_BUSINESS
	case SubscriptionLevelUltimate:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_ULTIMATE
	default:
		return users.SubscriptionLevel_SUBSCRIPTION_LEVEL_UNSPECIFIED
	}
}

// SubscriptionLevelFromProto преобразует protobuf SubscriptionLevel в доменный
func SubscriptionLevelFromProto(protoLevel users.SubscriptionLevel) SubscriptionLevel {
	switch protoLevel {
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_FREE:
		return SubscriptionLevelFree
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_BASIC:
		return SubscriptionLevelBasic
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_STANDARD:
		return SubscriptionLevelStandard
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_PRO:
		return SubscriptionLevelPro
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_PREMIUM:
		return SubscriptionLevelPremium
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_ENTERPRISE:
		return SubscriptionLevelEnterprise
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_LIFETIME:
		return SubscriptionLevelLifetime
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_STARTER:
		return SubscriptionLevelStarter
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_BUSINESS:
		return SubscriptionLevelBusiness
	case users.SubscriptionLevel_SUBSCRIPTION_LEVEL_ULTIMATE:
		return SubscriptionLevelUltimate
	default:
		return SubscriptionLevelUnspecified
	}
}
