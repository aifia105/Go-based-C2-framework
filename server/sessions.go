package server

import (
	"net"
	"reverse_shell/pkg/common"
	"reverse_shell/pkg/protocol"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Session struct {
	ID         string
	AgentID    string
	Conn       net.Conn
	Codec      *protocol.Codec
	Meta       map[string]string
	LastActive time.Time
	ResultChan chan protocol.Message
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSession(codec *protocol.Codec, conn net.Conn, agentID string, meta map[string]string) *Session {
	return &Session{
		ID:         common.GenerateID(),
		Conn:       conn,
		Codec:      codec,
		AgentID:    agentID,
		Meta:       meta,
		LastActive: time.Now(),
		ResultChan: make(chan protocol.Message, 10),
	}
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (manager *SessionManager) Add(session *Session) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.sessions[session.ID] = session
}

func (manager *SessionManager) Remove(sessionID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if session, exists := manager.sessions[sessionID]; exists {
		session.Conn.Close()
		if session.ResultChan != nil {
			close(session.ResultChan)
		}
		delete(manager.sessions, sessionID)
	}
}

func (manager *SessionManager) Get(sessionID string) (*Session, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	session, exists := manager.sessions[sessionID]
	return session, exists
}

func (manager *SessionManager) List() []*Session {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	sessions := make([]*Session, 0, len(manager.sessions))
	for _, session := range manager.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (manager *SessionManager) Touch(sessionID string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if session, exists := manager.sessions[sessionID]; exists {
		session.LastActive = time.Now()
	}
}

func (manager *SessionManager) StartCleanup(interval time.Duration, timeout time.Duration, logger *zap.Logger) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			manager.mu.Lock()
			now := time.Now()
			for id, session := range manager.sessions {
				if now.Sub(session.LastActive) > timeout {
					logger.Warn("Session timeout", zap.String("sessionID", id))
					session.Conn.Close()
					if session.ResultChan != nil {
						close(session.ResultChan)
					}
					delete(manager.sessions, id)
				}
			}
			manager.mu.Unlock()
		}
	}()
	return ticker
}
