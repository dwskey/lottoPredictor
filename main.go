// 로또 예측 프로그램 - 모듈화된 구조로 구성된 메인 파일
package main

import (
	"fmt"
	"log"
	"os"

	"lottopredictor/internal/analyzer"
	"lottopredictor/internal/config"
	"lottopredictor/internal/db"
	"lottopredictor/internal/fetcher"
	"lottopredictor/internal/output"
	"lottopredictor/internal/util"
)

func main() {
	config.LoadConfig("config.json")

	util.SeedCryptoRand() // 안전한 시드 초기화

	os.MkdirAll("database", os.ModePerm)
	database, err := db.InitDB("database/lotto.db")
	if err != nil {
		log.Fatalf("DB 초기화 실패: %v", err)
	}
	defer database.Close()

	latest := db.GetLatestDrawNumber(database)
	for i := latest + 1; ; i++ {
		result, err := fetcher.FetchDrawData(i)
		if err != nil {
			break
		}
		db.SaveDrawResult(database, result)
	}

	predictions := analyzer.Analyze(database)

	os.MkdirAll("result", os.ModePerm)
	output.SaveAsHTML(predictions, fmt.Sprintf("result/lotto_analysis_%d.html", predictions.DrawNumber))
	output.SaveAsTXT(predictions, fmt.Sprintf("result/lotto_analysis_%d.txt", predictions.DrawNumber))

}
