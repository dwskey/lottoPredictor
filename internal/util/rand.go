package util

import (
	"crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
)

// GlobalRand 안전한 시드로 초기화된 전역 rand 인스턴스
var GlobalRand *mathrand.Rand

// SeedCryptoRand 안전한 crypto 기반 시드로 전역 rand 생성
func SeedCryptoRand() {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic("crypto/rand 실패: " + err.Error())
	}
	seed := int64(binary.LittleEndian.Uint64(b[:]))
	GlobalRand = mathrand.New(mathrand.NewSource(seed))
}

// RandIntn returns a random integer in [0, n)
func RandIntn(n int) int {
	return GlobalRand.Intn(n)
}

// RandFloat64 returns a random float64 in [0.0, 1.0)
func RandFloat64() float64 {
	return GlobalRand.Float64()
}

// RandShuffle shuffles the provided slice using GlobalRand
func RandShuffle(slice []int) {
	GlobalRand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// RandRange returns a random integer in [min, max]
func RandRange(min, max int) int {
	if max < min {
		min, max = max, min // 범위가 반대면 교환
	}
	return min + GlobalRand.Intn(max-min+1)
}

// RandChoice returns one random element from a non-empty slice.
// Panics if the slice is empty.
func RandChoice(slice []int) int {
	if len(slice) == 0 {
		panic("RandChoice: empty slice")
	}
	return slice[GlobalRand.Intn(len(slice))]
}
