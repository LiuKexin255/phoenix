package main

import (
	"context"

	"phoenix/project/infra/object_storage/storage/proto"

	"google.golang.org/grpc/grpclog"
)

type server struct {
	proto.UnimplementedStoragerServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	grpclog.Infof("Received: %v", in.GetName())

	return &proto.HelloReply{
		Message: "Hello " + in.GetName(),
	}, nil
}
