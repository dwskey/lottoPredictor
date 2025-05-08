package test

import (
	"lottopredictor/internal/util" // 명시적으로 import 필요
	"testing"
)

func TestRandChoice(t *testing.T) {
	util.SeedCryptoRand()

	values := []int{1, 2, 3}
	val := util.RandChoice(values)

	if val != 1 && val != 2 && val != 3 {
		t.Errorf("RandChoice returned invalid value: %d", val)
	}
}
