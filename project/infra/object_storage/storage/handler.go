package main

import (
	"context"
	"fmt"

	"phoenix/common/go/otel"
	proto "phoenix/project/infra/object_storage"

	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func newServer() (*server, error) {
	sayHelloCounter, err := otel.Meter().Int64Counter("sayHello.counter",
		metric.WithUnit("1"),
		metric.WithDescription("hello say counter"),
	)
	if err != nil {
		return nil, err
	}

	return &server{
		sayHelloCounter: sayHelloCounter,
	}, nil
}

type server struct {
	proto.UnimplementedStoragerServer

	sayHelloCounter metric.Int64Counter
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	ctx, span := otel.Tracer().Start(ctx, "SayHello")
	defer span.End()

	s.sayHelloCounter.Add(ctx, 1)

	otel.Logger(ctx).Info(fmt.Sprintf("%s say hello", in.GetName()), zap.String("name", in.GetName()))

	return &proto.HelloReply{
		Message: "Hello " + in.GetName(),
	}, nil
}
