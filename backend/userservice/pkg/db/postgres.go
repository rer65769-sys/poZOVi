package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config - конфигурация PostgreSQL
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

// ConnectPostgres подключается к PostgreSQL
func ConnectPostgres(cfg PostgresConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}

// RunMigrations запускает миграции
func RunMigrations(db *sqlx.DB, migrationsDir string) error {
	// Используйте golang-migrate или другой инструмент для миграций
	// Пример с github.com/golang-migrate/migrate:
	// m, err := migrate.New(
	//     fmt.Sprintf("file://%s", migrationsDir),
	//     db.DriverName(),
	// )
	// if err != nil {
	//     return err
	// }
	// return m.Up()

	// Для простоты запустим SQL напрямую
	sql := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := db.Exec(sql)
	return err
}
