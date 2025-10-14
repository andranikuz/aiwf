package store

import (
	"crypto/sha1"
	"encoding/hex"
)

func hashBytes(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

func hashString(input string) string {
	return hashBytes([]byte(input))
}
