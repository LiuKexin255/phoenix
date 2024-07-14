package main

import (
	"context"
	"log"

	"phoenix/project/infra/object_storage/storage/proto"
)

type server struct {
	proto.UnimplementedStoragerServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())

	return &proto.HelloReply{
		Message: "Hello " + in.GetName(),
	}, nil
}
