package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
)

type HMAC struct {
	sha256Key string
}

func NewHMAC(sha256Key string) *HMAC {
	return &HMAC{
		sha256Key: sha256Key,
	}
}

func (h *HMAC) Create() hash.Hash {
	return hmac.New(sha256.New, []byte(h.sha256Key))
}
