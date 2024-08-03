package main

import (
	"context"
	"log"
	"time"

	proto "phoenix/project/infra/object_storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient("object-storage--storage-service.test:8888",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithKeepaliveParams(kacp),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewStoragerClient(conn)

	// Contact the server and print out its response.
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		r, err := c.SayHello(ctx, &proto.HelloRequest{Name: "liukexin"})
		if err != nil {
			log.Printf("could not greet: %v", err)
			continue
		}
		log.Printf("Greeting: %s", r.GetMessage())
	}
}
