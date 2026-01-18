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
	MsgExec   = "exec"
	MsgResult = "result"
	MsgPing   = "ping"
	MsgPong   = "pong"
	MsgError  = "error"
)
