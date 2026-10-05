package userservice

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

const (
	size = 32
)

var (
	b64Encoding = base64.StdEncoding
)

type SessionID [size]byte

func newSessionID() SessionID {
	b := make([]byte, size)
	rand.Read(b)
	return SessionID(b)
}

func newSessionIdFromB64(id string) (SessionID, error) {
	b := make([]byte, size)

	_, err := b64Encoding.Decode(b, []byte(id))
	if err != nil {
		return SessionID{}, err
	}

	return SessionID(b[:]), nil
}

func (id SessionID) Sha256() []byte {
	hashArray := sha256.Sum256(id[:])
	hashSlice := make([]byte, 32)
	copy(hashSlice, hashArray[:])
	return hashSlice
}

func (id SessionID) Base64() string {
	return b64Encoding.EncodeToString(id[:])
}

func (id SessionID) Base64Sha256() string {
	return b64Encoding.EncodeToString(id.Sha256())
}

func (id SessionID) HexSha256() string {
	return hex.EncodeToString(id.Sha256())
}
