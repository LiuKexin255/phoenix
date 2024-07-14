package main

import (
	"context"
	"log"
	"net/http"

	"phoenix/common/go/x509"
	"phoenix/project/infra/object_storage/storage/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

func main() {
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.NewClient(
		"127.0.0.1:8888",
		grpc.WithTransportCredentials(x509.MustGetCACert()),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = proto.RegisterStoragerHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}
	log.Fatalln(gwServer.ListenAndServe())
}
