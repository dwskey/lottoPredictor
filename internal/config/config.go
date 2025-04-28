package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	SuggestionSetCount int     `json:"suggestion_set_count"`
	LookbackRounds     int     `json:"lookback_rounds"`
	GAPBoostMultiplier float64 `json:"gap_boost_multiplier"` // 확률 계산에 영향 (보정 가중치)
	GapThreshold       int     `json:"gap_threshold"`        // 분석 통계에 영향 (미등장 번호 표시용)
}

var AppConfig Config

func LoadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("설정 파일 열기 실패: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		log.Fatalf("설정 파일 파싱 실패: %v", err)
	}
}
