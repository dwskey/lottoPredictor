// internal/util/rand.go
package util

import (
	"crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
)

// SeedCryptoRand 안전한 crypto 기반 시드로 math/rand 초기화
func SeedCryptoRand() {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic("crypto/rand 실패: " + err.Error())
	}
	mathrand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}
