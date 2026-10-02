package randx

import (
	cryptorand "crypto/rand"
)

type Generator interface {
	Int(low, high int64) int64
	IntUniqueN(low, high, n int64) []int64
}

func NewSeed() [32]byte {
	seed := make([]byte, 32)
	cryptorand.Read(seed)
	return [32]byte(seed)
}
