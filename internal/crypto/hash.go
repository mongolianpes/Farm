package crypto

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"io"
)

const (
	saltSize = 32
	iter     = 3
)

var mainSalt = []byte("?da;lmjgagma&qgmaf")

func HashString(str string) (string, error) {
	strByte := []byte(str)

	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	h := sha512.New()
	h.Write(salt)
	h.Write(mainSalt)

	h.Write(strByte)
	sum := h.Sum(nil)

	for i := 1; i < iter; i++ {
		h2 := sha512.New()
		h2.Write(sum)
		sum = h2.Sum(nil)
	}

	result := hex.EncodeToString(sum) + hex.EncodeToString(salt)

	return result, nil
}

func VerifyHash(candidateStr, hashAndSaltHex string) bool {
	salt, err := hex.DecodeString(hashAndSaltHex[128:])
	if err != nil {
		return false
	}
	expectedHash, err := hex.DecodeString(hashAndSaltHex[:128])
	if err != nil {
		return false
	}
	candidate := []byte(candidateStr)

	h := sha512.New()
	h.Write(salt)

	h.Write(mainSalt)
	h.Write(candidate)
	sum := h.Sum(nil)

	for i := 1; i < iter; i++ {
		h2 := sha512.New()
		h2.Write(sum)
		sum = h2.Sum(nil)
	}

	if len(sum) != len(expectedHash) {
		return false
	}
	return subtle.ConstantTimeCompare(sum, expectedHash) == 1
}
