package grpchandler

import (
	"context"
	hello "gateway/gen"
	"gateway/internal/domain"
	"log"
	"time"
)

type Handler struct {
	hello.UnimplementedGreeterServer
	service domain.HelloService
}

// NewHandler создает новый обработчик
func NewHandler(service domain.HelloService) *Handler {
	return &Handler{
		service: service,
	}
}

// SayHello обрабатывает gRPC вызов
func (h *Handler) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloReply, error) {
	// Логируем входящий запрос
	log.Printf("Received gRPC request for name: %s", req.GetName())

	// Добавляем небольшую задержку для имитации работы
	time.Sleep(100 * time.Millisecond)

	// Преобразование protobuf в доменную модель
	domainReq := &domain.HelloRequest{
		Name: req.GetName(),
	}

	// Вызов доменного сервиса
	resp, err := h.service.SayHello(domainReq)
	if err != nil {
		log.Printf("Service error: %v", err)
		return nil, err
	}

	// Преобразование доменной модели в protobuf
	return &hello.HelloReply{
		Message: resp.Message,
	}, nil
}
