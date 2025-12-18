package httpgateway

import (
	"context"
	"fmt"
	hello "gateway/gen"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/proto"
)

type GRPCConnectionManager struct {
	mu         sync.RWMutex
	conn       *grpc.ClientConn
	addr       string
	retries    int
	maxRetries int
}

// Gateway представляет HTTP Gateway
type Gateway struct {
	grpcMgr  *GRPCConnectionManager
	router   *mux.Router
	httpAddr string
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewGRPCConnectionManager создает новый менеджер соединений
func NewGRPCConnectionManager(addr string) *GRPCConnectionManager {
	mgr := &GRPCConnectionManager{
		addr:       addr,
		maxRetries: 5,
	}
	if err := mgr.connect(); err != nil {
		log.Printf("Initial connection failed: %v", err)
	}
	return mgr
}

func (m *GRPCConnectionManager) connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Закрываем старое соединение если есть
	if m.conn != nil {
		m.conn.Close()
	}

	// Настраиваем keepalive для поддержания соединения
	kacp := keepalive.ClientParameters{
		Time:                30 * time.Second, // Увеличили для тестирования
		Timeout:             5 * time.Second,  // Увеличили таймаут
		PermitWithoutStream: true,
	}

	// Увеличиваем максимальный размер сообщения
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(),                   // Блокируем до установления соединения
		grpc.WithTimeout(10 * time.Second), // Таймаут на подключение
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, m.addr, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	m.conn = conn
	m.retries = 0
	log.Printf("Connected to gRPC server at %s", m.addr)
	return nil
}

func (m *GRPCConnectionManager) GetConnection() (*grpc.ClientConn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.conn == nil {
		return nil, fmt.Errorf("no connection available")
	}

	state := m.conn.GetState()
	log.Printf("gRPC connection state: %v", state)

	return m.conn, nil
}

func (m *GRPCConnectionManager) ensureConnection() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.retries >= m.maxRetries {
		log.Printf("Max retries reached for gRPC connection")
		return
	}

	m.retries++
	log.Printf("Attempting to reconnect to gRPC server (attempt %d/%d)",
		m.retries, m.maxRetries)

	// Экспоненциальная задержка
	delay := time.Duration(1<<uint(m.retries)) * time.Second
	time.Sleep(delay)

	if err := m.connect(); err != nil {
		log.Printf("Reconnect failed: %v", err)
	}
}

func (m *GRPCConnectionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.conn != nil {
		m.conn.Close()
		log.Printf("gRPC connection closed")
	}
}

// NewGateway создает новый Gateway
func NewGateway(grpcAddr, httpAddr string) *Gateway {
	ctx, cancel := context.WithCancel(context.Background())

	return &Gateway{
		grpcMgr:  NewGRPCConnectionManager(grpcAddr),
		router:   mux.NewRouter(),
		httpAddr: httpAddr,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SetupRoutes настраивает маршруты
func (g *Gateway) SetupRoutes() error {
	// Создаем мультиплексор gRPC Gateway с настройками
	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(g.errorHandler),
		runtime.WithForwardResponseOption(g.responseModifier),
	)

	// Получаем соединение
	_, err := g.grpcMgr.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get gRPC connection: %w", err)
	}

	// Настраиваем параметры для gRPC Gateway
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
	}

	// Регистрируем gRPC Gateway
	err = hello.RegisterGreeterHandlerFromEndpoint(g.ctx, gwmux, g.grpcMgr.addr, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	// Основные маршруты
	g.router.HandleFunc("/", g.homeHandler)
	g.router.HandleFunc("/health", g.healthHandler)

	// Все запросы к /v1/ передаем в gRPC Gateway
	g.router.PathPrefix("/v1/").Handler(gwmux)

	return nil
}

// errorHandler обрабатывает ошибки gRPC
func (g *Gateway) errorHandler(ctx context.Context, mux *runtime.ServeMux,
	marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("gRPC Gateway error: %v", err)

	// Если ошибка связана с соединением, пытаемся переподключиться
	if err.Error() == "grpc: the client connection is closing" {
		go g.grpcMgr.ensureConnection()
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "gRPC connection closed, reconnecting..."}`))
		return
	}

	// Стандартная обработка ошибок
	runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
}

// responseModifier модифицирует ответы
func (g *Gateway) responseModifier(ctx context.Context, w http.ResponseWriter,
	resp proto.Message) error {

	// Добавляем заголовки
	w.Header().Set("X-Gateway-Version", "1.0")
	return nil
}

// homeHandler обрабатывает главную страницу
func (g *Gateway) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<h1>gRPC Gateway (Clean Architecture)</h1>
		<p>Available endpoints:</p>
		<ul>
			<li><a href="/v1/hello/world">GET /v1/hello/world</a></li>
			<li>POST /v1/hello with JSON: {"name": "world"}</li>
			<li><a href="/health">GET /health</a></li>
		</ul>
	`)
}

// healthHandler обрабатывает проверку здоровья
func (g *Gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := g.grpcMgr.GetConnection()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status": "unavailable", "grpc": "disconnected", "error": "%v"}`, err)
		return
	}

	state := conn.GetState()
	log.Printf("Health check - gRPC state: %v", state)

	w.Header().Set("Content-Type", "application/json")
	if state == connectivity.Ready || state == connectivity.Idle {
		fmt.Fprintf(w, `{"status": "ok", "grpc_state": "%s"}`, state)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status": "degraded", "grpc_state": "%s"}`, state)
	}
}

// Run запускает HTTP сервер
func (g *Gateway) Run() error {
	server := &http.Server{
		Addr:         g.httpAddr,
		Handler:      g.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

// Close закрывает соединения
func (g *Gateway) Close() {
	g.cancel()
	g.grpcMgr.Close()
}
