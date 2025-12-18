package grpch

import (
	"context"
	"log"
	"time"
	users "userservice/gen/v1"
	"userservice/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	users.UnimplementedUserServiceServer
	service domain.UserService
}

func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *users.CreateUserRequest) (*users.User, error) {
	log.Printf("CreateUser request: %s", req.GetEmail())

	// Преобразуем protobuf запрос в доменную модель
	domainUser := domain.CreateUserRequestFromProto(req)

	// Регистрируем пользователя
	user, err := h.service.Register(domainUser)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return user.ToProto(), nil
}

func (h *UserHandler) GetUserById(ctx context.Context, req *users.GetUserByIdRequest) (*users.User, error) {
	log.Printf("GetUserById request for ID: %s", req.GetId())

	user, err := h.service.GetUser(req.GetId())
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if err == domain.ErrUserBanned {
			return nil, status.Error(codes.PermissionDenied, "user is banned")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *users.GetUserByEmailRequest) (*users.User, error) {
	log.Printf("GetUserByEmail request for email: %s", req.GetEmail())

	// Используем GetUser для консистентности
	user, err := h.service.GetUser(req.GetEmail())
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *users.UpdateUserRequest) (*users.User, error) {
	log.Printf("UpdateUser request for ID: %s", req.GetId())

	// Получаем текущего пользователя
	user, err := h.service.GetUser(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Обновляем поля из запроса
	user.UpdateFromProto(req)

	updatedUser, err := h.service.UpdateUser(user)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return updatedUser.ToProto(), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *users.DeleteUserRequest) (*emptypb.Empty, error) {
	log.Printf("DeleteUser request for ID: %s", req.GetId())

	if err := h.service.DeleteUser(req.GetId()); err != nil {
		if err == domain.ErrUserNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *users.ListUsersRequest) (*users.ListUsersResponse, error) {
	log.Printf("ListUsers request: page=%d, page_size=%d", req.GetPage(), req.GetPageSize())

	filter := &domain.UserFilter{
		Page:     int(req.GetPage()),
		PageSize: int(req.GetPageSize()),
		Search:   req.GetSearch(),
		Status:   domain.UserStatusFromProto(req.GetStatus()),
		Role:     domain.UserRoleFromProto(req.GetRole()),
	}

	if req.GetIsBanned() {
		isBanned := true
		filter.IsBanned = &isBanned
	}

	if req.GetSubscriptionStatus() != users.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED {
		subStatus := domain.SubscriptionStatusFromProto(req.GetSubscriptionStatus())
		filter.SubStatus = &subStatus
	}

	if req.GetSubscriptionLevel() != users.SubscriptionLevel_SUBSCRIPTION_LEVEL_UNSPECIFIED {
		subLevel := domain.SubscriptionLevelFromProto(req.GetSubscriptionLevel())
		filter.SubLevel = &subLevel
	}

	users, total, err := h.service.ListUsers(filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return domain.ListUsersResponseToProto(users, total, req.GetPage(), req.GetPageSize()), nil
}

func (h *UserHandler) Authenticate(ctx context.Context, req *users.AuthenticateRequest) (*users.AuthenticateResponse, error) {
	log.Printf("Authenticate request for email: %s", req.GetEmail())

	user, token, err := h.service.Authenticate(req.GetEmail(), req.GetPassword())
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		if err == domain.ErrUserBanned {
			return nil, status.Error(codes.PermissionDenied, "user is banned")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &users.AuthenticateResponse{
		Token:     token,
		User:      user.ToProto(),
		ExpiresAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}, nil
}

func (h *UserHandler) ValidateToken(ctx context.Context, req *users.ValidateTokenRequest) (*users.ValidateTokenResponse, error) {
	log.Printf("ValidateToken request")

	user, err := h.service.ValidateToken(req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &users.ValidateTokenResponse{
		Valid: true,
		User:  user.ToProto(),
	}, nil
}

func (h *UserHandler) BanUser(ctx context.Context, req *users.BanUserRequest) (*users.User, error) {
	log.Printf("BanUser request for user: %s", req.GetUserId())

	var duration *time.Duration
	if req.GetBannedUntil() != nil {
		dur := time.Until(req.GetBannedUntil().AsTime())
		duration = &dur
	}

	user, err := h.service.BanUser(req.GetUserId(), req.GetReason(), req.GetBannedBy(), duration)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) UnbanUser(ctx context.Context, req *users.UnbanUserRequest) (*users.User, error) {
	log.Printf("UnbanUser request for user: %s", req.GetUserId())

	user, err := h.service.UnbanUser(req.GetUserId(), req.GetUnbannedBy())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) UpdateSubscription(ctx context.Context, req *users.UpdateSubscriptionRequest) (*users.User, error) {
	log.Printf("UpdateSubscription request for user: %s", req.GetUserId())

	subscription := &domain.SubscriptionInfo{
		Status:            domain.SubscriptionStatusFromProto(req.GetStatus()),
		Level:             domain.SubscriptionLevelFromProto(req.GetLevel()),
		SubscriptionStart: time.Now(),
		SubscriptionID:    req.GetSubscriptionId(),
		PaymentMethod:     req.GetPaymentMethod(),
		AutoRenew:         req.GetAutoRenew(),
		Amount:            req.GetAmount(),
		Currency:          req.GetCurrency(),
		// Features:          req.GetFeatures(),
	}

	if req.GetSubscriptionEnd() != nil {
		end := req.GetSubscriptionEnd().AsTime()
		subscription.SubscriptionEnd = &end
	}

	if req.GetTrialEnd() != nil {
		trialEnd := req.GetTrialEnd().AsTime()
		subscription.TrialEnd = &trialEnd
	}

	// if req.GetNextBillingDate() != nil {
	// 	nextBilling := req.GetNextBillingDate().AsTime()
	// 	subscription.NextBillingDate = &nextBilling
	// }

	user, err := h.service.UpdateSubscription(req.GetUserId(), subscription)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) CancelSubscription(ctx context.Context, req *users.CancelSubscriptionRequest) (*users.User, error) {
	log.Printf("CancelSubscription request for user: %s", req.GetUserId())

	user, err := h.service.CancelSubscription(req.GetUserId(), req.GetReason(), req.GetImmediateCancellation())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return user.ToProto(), nil
}

func (h *UserHandler) HealthCheck(ctx context.Context, req *users.HealthCheckRequest) (*users.HealthCheckResponse, error) {
	log.Printf("HealthCheck request")

	healthy, err := h.service.HealthCheck()
	if err != nil || !healthy {
		return nil, status.Error(codes.Internal, "service unhealthy")
	}

	return &users.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "user-service",
		Version:   "1.0.0",
	}, nil
}
