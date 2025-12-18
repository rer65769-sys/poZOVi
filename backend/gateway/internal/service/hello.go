package service

import "gateway/internal/domain"

type helloService struct{}

func NewHelloService() domain.HelloService {
	return &helloService{}
}

func (s *helloService) SayHello(req *domain.HelloRequest) (*domain.HelloResponse, error) {
	return &domain.HelloResponse{
		Message: "Hello, " + req.Name + "!",
	}, nil
}
