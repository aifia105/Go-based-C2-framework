package server

import (
	"errors"
	"net"
	"os"
	"reverse_shell/pkg/common"
	"reverse_shell/pkg/protocol"
	"time"

	"go.uber.org/zap"
)

func HandleNewConnection(conn net.Conn, sessionManager *SessionManager, logger *zap.Logger) error {
	codec := protocol.NewCodec(conn)

	authMsg, err := codec.Read()
	if err != nil {
		logger.Error("Auth handshake failed", zap.Error(err))
		conn.Close()
		return err
	}

	if authMsg.Type != protocol.MsgAuth {
		logger.Error("Invalid handshake message type", zap.String("type", authMsg.Type))
		conn.Close()
		return errors.New("invalid handshake message type")
	}

	if !validateAuthPayload(authMsg.Payload) {
		logger.Error("Authentication failed", zap.String("agentID", authMsg.ID))
		conn.Close()
		return errors.New("authentication failed")
	}

	session := NewSession(codec, conn, authMsg.ID, authMsg.Meta)
	sessionManager.Add(session)

	authResponse := protocol.Message{
		Type:      protocol.MsgAuth,
		Payload:   "auth_success",
		ID:        session.AgentID,
		SessionID: session.ID,
	}
	if err := codec.Send(authResponse); err != nil {
		logger.Error("Failed to send auth ack", zap.String("sessionID", session.ID), zap.Error(err))
		conn.Close()
		return err
	}

	logger.Info("New session created", zap.String("sessionID", session.ID), zap.String("agentID", session.AgentID))

	go func() {
		defer sessionManager.Remove(session.ID)
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
			msg, err := codec.Read()
			if err != nil {
				logger.Error("Error reading message", zap.String("sessionID", session.ID), zap.Error(err))
				return
			}
			logger.Info("Message received", zap.String("sessionID", session.ID), zap.String("messageType", msg.Type))
			sessionManager.Touch(session.ID)

			switch msg.Type {
			case protocol.MsgResult:
				logger.Info("Command output received", zap.String("sessionID", session.ID), zap.String("output", msg.Payload))
			case protocol.MsgError:
				logger.Error("Error message from agent", zap.String("sessionID", session.ID), zap.String("error", msg.Payload))
			case protocol.MsgPong:
				logger.Info("Pong received", zap.String("sessionID", session.ID))
			default:
				logger.Warn("Unknown message type", zap.String("sessionID", session.ID), zap.String("messageType", msg.Type))
			}
		}
	}()
	return nil

}

func validateAuthPayload(payload string) bool {
	expectedAuth := os.Getenv("AGENT_AUTH_FLAG")
	if expectedAuth == "" {
		return false
	}
	return payload == expectedAuth
}

func SendCommand(sessionID, command string, sessionManager *SessionManager, logger *zap.Logger) error {
	session, exists := sessionManager.Get(sessionID)
	if !exists {
		logger.Error("Session not found", zap.String("sessionID", sessionID))
		return errors.New("session not found")
	}

	msg := protocol.Message{
		Type:      protocol.MsgExec,
		SessionID: sessionID,
		ID:        common.GenerateID(),
		Payload:   command,
	}

	err := session.Codec.Send(msg)
	if err != nil {
		logger.Error("Error sending message to session", zap.String("sessionID", sessionID), zap.Error(err))
		return err
	}
	logger.Info("Command sent", zap.String("sessionID", sessionID), zap.String("command", command))
	return nil
}
