package main

import (
	httpgateway "gateway/internal/delivery/http"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Создаем и настраиваем Gateway
	gateway := httpgateway.NewGateway("localhost:50051", ":8888")

	// Настраиваем маршруты
	if err := gateway.SetupRoutes(); err != nil {
		log.Fatalf("Failed to setup routes: %v", err)
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем Gateway в горутине
	go func() {
		log.Printf("Starting HTTP Gateway on :8888")
		if err := gateway.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run gateway: %v", err)
		}
	}()

	// Ждем сигнала для завершения
	<-stop
	log.Println("Shutting down gateway...")

	// Закрываем соединения
	gateway.Close()
	log.Println("Gateway shutdown complete")
}
