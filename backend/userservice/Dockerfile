# Build stage
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для protoc
RUN apk add --no-cache protoc protobuf-dev git

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем gRPC код
RUN mkdir -p third_party && \
    git clone https://github.com/googleapis/googleapis.git third_party/googleapis && \
    chmod +x scripts/generate.sh && \
    ./scripts/generate.sh

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder stage
COPY --from=builder /user-service .

# Копируем конфигурацию
COPY config ./config

# Копируем сгенерированный код
COPY --from=builder /app/gen ./gen

EXPOSE 50051 8080

CMD ["./user-service"]