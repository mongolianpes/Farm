package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func RandHex32() (string, error) {
	currentNumBytes := 32
	id := make([]byte, currentNumBytes)
	numBytes, err := rand.Read(id)
	if err != nil {
		return "", err
	}

	if numBytes != currentNumBytes {
		return "", errors.New("Few bytes used")
	}

	return hex.EncodeToString(id), nil
}
