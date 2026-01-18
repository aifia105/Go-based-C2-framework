package crypto_tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

func TLSClient(certFile, serverName string) (*tls.Config, error) {
	cert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(cert); !ok {
		return nil, errors.New("failed to append certificate")
	}
	return &tls.Config{
		RootCAs:    caPool,
		MinVersion: tls.VersionTLS13,
		ServerName: serverName,
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
