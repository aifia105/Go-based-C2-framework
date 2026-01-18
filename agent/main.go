package agent

import (
	"log"
)

func main() {
	serverName := "server.local"
	addr := "127.0.0.1:8080"
	caFile := "ca.crt"

	client, err := Run(addr, caFile, serverName)
	if err != nil {
		log.Fatal(err)
	}
}
