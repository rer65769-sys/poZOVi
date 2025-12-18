package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	GRPC     GRPCConfig
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Mongo    MongoConfig
	JWT      JWTConfig
	Redis    RedisConfig
	Log      LogConfig
}

type AppConfig struct {
	Name    string
	Env     string
	Version string
}

type GRPCConfig struct {
	Port int
}

type HTTPConfig struct {
	Port int
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int
	MaxIdle  int
}

type MongoConfig struct {
	URI         string
	Database    string
	MaxPoolSize uint64
	MinPoolSize uint64
	Timeout     time.Duration
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type LogConfig struct {
	Level  string
	Format string
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Установка значений по умолчанию
	viper.SetDefault("app.name", "user-service")
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("grpc.port", 50051)
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.user", "user")
	viper.SetDefault("postgres.password", "password")
	viper.SetDefault("postgres.name", "users")
	viper.SetDefault("postgres.sslmode", "disable")
	viper.SetDefault("postgres.max_conns", 50)
	viper.SetDefault("postgres.max_idle", 10)
	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "userservice")
	viper.SetDefault("mongo.max_pool_size", 100)
	viper.SetDefault("mongo.min_pool_size", 10)
	viper.SetDefault("mongo.timeout", "10s")
	viper.SetDefault("jwt.expiry", "24h")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	// Чтение конфигурации
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}
