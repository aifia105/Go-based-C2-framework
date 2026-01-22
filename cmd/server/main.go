package main

import (
	"flag"
	"os"
	"os/signal"
	"reverse_shell/server"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Server started", zap.String("version", "1.0.0"))

	addr := flag.String("addr", "", "Listen address")
	certFile := flag.String("cert", "", "Server certificate file")
	keyFile := flag.String("key", "", "Server key file")
	flag.Parse()

	_, exist := os.LookupEnv("AGENT_AUTH_FLAG")
	if !exist {
		logger.Fatal("AGENT_AUTH_FLAG environment variable not set")
	}

	sessionManager := server.NewSessionManager()

	ticker := sessionManager.StartCleanup(5*time.Minute, 10*time.Minute, logger)
	defer ticker.Stop()

	stopChan := make(chan struct{})
	listenerErrChan := make(chan error, 1)

	go func() {
		err := server.StartListener(*addr, *certFile, *keyFile, sessionManager, logger, stopChan)
		listenerErrChan <- err
	}()

	go server.RunCLI(sessionManager, logger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Received shutdown signal")
	case err := <-listenerErrChan:
		logger.Error("Listener error", zap.Error(err))

	}
	close(stopChan)
	logger.Info("Server stopped")
}
