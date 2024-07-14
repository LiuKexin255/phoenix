package main

import (
	"context"
	"log"
	"time"

	"phoenix/common/go/x509"
	"phoenix/project/infra/object_storage/storage/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient(
		"127.0.0.1:8888",
		grpc.WithTransportCredentials(x509.MustGetCACert()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewStoragerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &proto.HelloRequest{Name: "liukexin"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
