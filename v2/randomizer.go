package toolkit

import (
	crand "crypto/rand"
	"math/big"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		x, _ := crand.Int(crand.Reader, big.NewInt(int64(len(r))))
		s[i] = r[x.Uint64()]
	}
	return string(s)
}
