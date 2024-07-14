package main

import (
	"log"
	"net"

	"phoenix/project/infra/object_storage/storage/proto"
	"phoenix/common/go/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(x509.MustGetServerCert())),
	)

	proto.RegisterStoragerServer(s, new(server))

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
