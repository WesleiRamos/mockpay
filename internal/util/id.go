package util

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	ts := time.Now().UTC().Format("060102150405")
	return ts + hex.EncodeToString(b)
}
