package protocol

type Message struct {
	Type      string            `json:"type"`
	SessionID string            `json:"session_id,omitempty"`
	ID        string            `json:"id,omitempty"`
	Payload   string            `json:"payload,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

const (
	MsgHello  = "hello"
	MsgAuth   = "auth"
	MsgExec   = "exec"
	MsgResult = "result"
	MsgPing   = "ping"
	MsgPong   = "pong"
	MsgError  = "error"
)

const (
	ProtocolVersion = "1.0"
	MaxMessageSize  = 10 * 1024 * 1024
)
