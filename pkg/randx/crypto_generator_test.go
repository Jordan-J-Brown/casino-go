package randx

import (
	"encoding/hex"
	"fmt"
	"slices"
	"testing"
)

func Test(t *testing.T) {
	gen := NewCryptoGenerator()
	nums := gen.IntUniqueN(0, 24, 24)
	slices.Sort(nums)
	fmt.Println(nums)
	fmt.Println(hex.EncodeToString(gen.Seed[:]))
}
