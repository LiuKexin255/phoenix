package main

import (
	"context"

	"phoenix/project/infra/object_storage/storage/proto"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/grpclog"
)

func newServer(tracer trace.Tracer, meter metric.Meter) (*server, error) {
	sayHelloCounter, err := meter.Int64Counter("sayHello.counter",
		metric.WithUnit("1"),
		metric.WithDescription("hello say counter"),
	)
	if err != nil {
		return nil, err
	}

	return &server{
		tracer: tracer,
		meter:  meter,

		sayHelloCounter: sayHelloCounter,
	}, nil
}

type server struct {
	proto.UnimplementedStoragerServer

	tracer trace.Tracer
	meter  metric.Meter

	sayHelloCounter metric.Int64Counter
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *proto.HelloRequest) (*proto.HelloReply, error) {
	ctx, span := s.tracer.Start(ctx, "SayHello")
	defer span.End()

	grpclog.Infof("Received: %v", in.GetName())

	return &proto.HelloReply{
		Message: "Hello " + in.GetName(),
	}, nil
}
