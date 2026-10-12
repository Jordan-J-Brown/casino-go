package casinosession

import (
	"math/rand"
	"strconv"
)

func generateUsername() string {
	return strconv.FormatUint(uint64(rand.Uint32()), 10)
}
