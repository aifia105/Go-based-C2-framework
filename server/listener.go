package server

import (
	"crypto/tls"
	"reverse_shell/pkg/crypto_tls"

	"go.uber.org/zap"
)

func StartListener(addr string, certFile, keyFile string, sessionManager *SessionManager, logger *zap.Logger, stopChan <-chan struct{}) error {
	config, err := crypto_tls.TLSServer(certFile, keyFile)
	if err != nil {
		return err
	}

	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}

	logger.Info("Server listening", zap.String("address", addr))

	go func() {
		<-stopChan
		logger.Info("Stopping listener...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-stopChan:
				return nil
			default:
				logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
		}

		logger.Info("New connection", zap.String("remote", conn.RemoteAddr().String()))

		if tlsConn, ok := conn.(*tls.Conn); ok {
			if err := tlsConn.Handshake(); err != nil {
				logger.Error("TLS handshake failed", zap.Error(err))
				conn.Close()
				continue
			}
		}

		go HandleNewConnection(conn, sessionManager, logger)
	}
}
