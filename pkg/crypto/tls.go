package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

func NewTLSClient(certFile, serverName string) (*tls.Config, error) {
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("failed to parse CA certificate")
	}
	return &tls.Config{
		ServerName: serverName,
		RootCAs:    caPool,
		MinVersion: tls.VersionTLS13,
	}, nil
}
func TLSServer(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}
