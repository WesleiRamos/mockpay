package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
