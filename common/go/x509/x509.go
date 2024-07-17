package x509

import (
	"crypto/tls"
	_ "embed"
	"log"
	"path"

	"google.golang.org/grpc/credentials"
)

const (
	defaultGRPCTLSPath = "/etc/tls/grpc"
	defaultTLSCertFile = "tls.crt"
	defaultTLSKeyFile  = "tls.key"

	defaultCACertPath = "/etc/ca/grpc"
	defaultCACertFile = "ca.crt"

	grpcHost = "grpc.liukexin.com"
)

// MustGetServerCert 获取服务证书
func MustGetServerCert() *tls.Certificate {
	c, err := tls.LoadX509KeyPair(path.Join(defaultGRPCTLSPath, defaultTLSCertFile), path.Join(defaultGRPCTLSPath, defaultTLSKeyFile))
	if err != nil {
		log.Panic(err)
	}
	return &c
}

// MustGetCACert 获取根证书
func MustGetCACert() credentials.TransportCredentials {
	c, err := credentials.NewClientTLSFromFile(path.Join(defaultCACertPath, defaultCACertFile), grpcHost)
	if err != nil {
		log.Panic(err)
	}
	return c
}
