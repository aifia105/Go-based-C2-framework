package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return fmt.Sprintf("error-generating-id-%v", err)
	}
	return hex.EncodeToString(bytes)
}
