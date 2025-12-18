package grpcserver

import (
	hello "gateway/gen"
	grpchandler "gateway/internal/delivery/grpc"
	"log"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	addr   string
}

// NewServer создает новый gRPC сервер
func NewServer(handler *grpchandler.Handler, addr string) *Server {
	s := grpc.NewServer()
	hello.RegisterGreeterServer(s, handler)

	return &Server{
		server: s,
		addr:   addr,
	}
}

// Run запускает gRPC сервер
func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	log.Printf("gRPC server listening at %v", s.addr)
	return s.server.Serve(lis)
}

// GracefulStop останавливает сервер
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}
