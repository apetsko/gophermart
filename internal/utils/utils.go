package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func GenerateID(s string, length int) (id string) {
	hash := sha256.Sum256([]byte(s))
	id = base64.RawURLEncoding.EncodeToString(hash[:length])[:length]
	return
}

func GenerateUserID(length int) (id string, err error) {
	r := make([]byte, length)

	_, err = rand.Read(r)
	if err != nil {
		err = fmt.Errorf("failed to generate random User ID: %w", err)
		return "", err
	}

	id = hex.EncodeToString(r)

	return id, nil
}
