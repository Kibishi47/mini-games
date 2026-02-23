package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateToken() (rawToken string, tokenHash string, err error) {
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		return "", "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(b)
	tokenHash = HashToken(rawToken)

	return rawToken, tokenHash, nil
}

func HashToken(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
