package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	users "userservice/gen/v1"
	"userservice/internal/config"
	"userservice/internal/delivery/grpch"
	"userservice/internal/repository/mongodb"
	"userservice/internal/repository/postgres"
	"userservice/internal/server"
	"userservice/pkg/db"
	"userservice/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к PostgreSQL
	postgresDB, err := db.ConnectPostgres(db.PostgresConfig(cfg.Postgres))
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Подключение к MongoDB
	mongoClient, err := db.ConnectMongoDB(db.MongoConfig(cfg.Mongo))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Инициализация репозиториев
	postgresRepo := postgres.NewPostgresUserRepository(postgresDB)
	mongoRepo := mongodb.NewMongoUserRepository(mongoClient, cfg.Mongo.Database)

	// Инициализация JWT менеджера
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Инициализация сервиса
	userService := server.NewUserService(postgresRepo, mongoRepo, jwtManager, cfg)

	// Создание gRPC обработчика
	userHandler := grpch.NewUserHandler(userService)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*10), // 10MB
		grpc.MaxSendMsgSize(1024*1024*10), // 10MB
	)

	// Регистрация сервисов
	users.RegisterUserServiceServer(grpcServer, userHandler)

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Reflection для тестирования
	if cfg.App.Env == "development" {
		reflection.Register(grpcServer)
	}

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Starting gRPC server on port %d", cfg.GRPC.Port)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Ждем сигнала завершения
	<-stop
	log.Println("Shutting down gRPC server...")

	// Устанавливаем статус NOT_SERVING для health check
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Graceful stop с таймаутом
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	log.Println("gRPC server stopped gracefully")
}
