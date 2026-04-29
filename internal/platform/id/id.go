package id

import (
	"crypto/rand"
	"encoding/hex"
)

func New(prefix string) string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)

	return prefix + "_" + hex.EncodeToString(bytes)
}
