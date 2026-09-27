package randx

import (
	"bytes"
	"encoding/gob"
	"math/rand/v2"
	"sync"
)

type CryptoGenerator struct {
	_    sync.Mutex //prevent copying
	r    *rand.Rand
	Seed [32]byte
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
		Seed: seed,
	}
}

func (gen *CryptoGenerator) Int(low, high int64) int64 {
	return gen.r.Int64N((high+1)-low) + low
}

func (gen *CryptoGenerator) IntUniqueN(low, high, n int64) []int64 {
	set := map[int64]struct{}{}
	out := make([]int64, 0, n)
	for len(out) != int(n) {
		rn := gen.Int(low, high)
		if _, exists := set[rn]; !exists {
			set[rn] = struct{}{}
			out = append(out, rn)
		}
	}
	return out
}

func init() {
	gob.Register(&CryptoGenerator{})
}

type cryptoGeneratorAlias CryptoGenerator

func (gen *CryptoGenerator) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode((*cryptoGeneratorAlias)(gen)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (gen *CryptoGenerator) UnmarshalBinary(data []byte) error {
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode((*cryptoGeneratorAlias)(gen)); err != nil {
		return err
	}

	gen.r = rand.New(rand.NewChaCha8(gen.Seed))

	return nil
}
