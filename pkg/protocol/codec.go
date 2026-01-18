package protocol

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
)

type Codec struct {
	conn net.Conn
}

func NewCodec(conn net.Conn) *Codec {
	return &Codec{conn: conn}
}

func EncodeJSON(message Message) ([]byte, error) {
	return json.Marshal(message)
}

func DecodeJSON(data []byte) (Message, error) {
	var message Message
	err := json.Unmarshal(data, &message)
	return message, err
}

func (codec *Codec) Send(message Message) error {
	data, err := EncodeJSON(message)
	if err != nil {
		return err
	}
	length := uint32(len(data))
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, length)
	_, err = codec.conn.Write(prefix)
	if err != nil {
		return err
	}
	_, err = codec.conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (codec *Codec) Read() (Message, error) {
	var message Message
	prefix := make([]byte, 4)
	if _, err := io.ReadFull(codec.conn, prefix); err != nil {
		return message, err
	}
	length := binary.BigEndian.Uint32(prefix)
	if length == 0 {
		return message, errors.New("empty message")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(codec.conn, data); err != nil {
		return message, err
	}
	if length != uint32(len(data)) {
		return message, errors.New("invalid message length")
	}
	message, err := DecodeJSON(data)
	if err != nil {
		return message, err
	}
	return message, nil
}
