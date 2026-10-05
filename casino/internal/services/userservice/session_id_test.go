package userservice

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	id := newSessionID()
	b64 := id.Base64()
	b64Hash := id.Base64Sha256()
	hexHash := id.HexSha256()
	fmt.Println(b64Hash, len(b64Hash))
	fmt.Println(b64, len(b64))
	fmt.Println(hexHash, len(hexHash))
}
