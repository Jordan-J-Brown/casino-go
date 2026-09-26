package randx

import (
	cryptorand "crypto/rand"
	"math/rand/v2"
	"sync"
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

type CryptoGenerator struct {
	_    sync.Mutex //prevent copying
	r    *rand.Rand
	seed [32]byte
}

func NewCryptoGenerator() *CryptoGenerator {
	return newCryptoGenerator(NewSeed())
}

func NewCryptoGeneratorFromSeed(seed [32]byte) *CryptoGenerator {
	return newCryptoGenerator(seed)
}

func newCryptoGenerator(seed [32]byte) *CryptoGenerator {
	r := rand.New(rand.NewChaCha8([32]byte(seed)))
	return &CryptoGenerator{
		r:    r,
		seed: seed,
	}
}

func (gen *CryptoGenerator) Int(low, high int64) int64 {
	return gen.r.Int64N(high-low) + low
}

func (gen *CryptoGenerator) IntUniqueN(low, high int64, n int) []int64 {
	set := map[int64]struct{}{}
	out := make([]int64, 0, n)
	for len(out) != n {
		rn := gen.r.Int64N(high-low) + low
		if _, exists := set[rn]; !exists {
			out = append(out, rn)
		}
	}
	return out
}
