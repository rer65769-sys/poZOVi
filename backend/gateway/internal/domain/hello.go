package domain

type HelloRequest struct {
	Name string
}

type HelloResponse struct {
	Message string
}

type HelloService interface {
	SayHello(req *HelloRequest) (*HelloResponse, error)
}
