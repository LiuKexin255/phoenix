package x509

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"log"

	"google.golang.org/grpc/credentials"
)

var (
	//go:embed server.cer
	serverCert []byte

	//go:embed server.key
	serverKey []byte

	//go:embed ca.cer
	caCert []byte
)

const (
	defaultGRPCHost = "grpc.liukexin.com"
)

// MustGetServerCert 获取服务证书
func MustGetServerCert() *tls.Certificate {
	c, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		log.Panic(err)
	}
	return &c
}

// MustGetCACert 获取根证书
func MustGetCACert() credentials.TransportCredentials {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(caCert) {
		log.Panic("credentials: failed to append certificates")
	}

	return credentials.NewClientTLSFromCert(cp, defaultGRPCHost)
}
