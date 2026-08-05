package sqlite

import (
	"crypto/rand"
	"encoding/base64"
)

func newID() string {
	return randomToken(16)
}

func newAPIKey() string {
	return randomToken(24)
}

func randomToken(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
