package test

import (
	"fmt"
	"log"
	"testing"

	"lottopredictor/internal/analyzer"
	"lottopredictor/internal/config"
	"lottopredictor/internal/db"
	"lottopredictor/internal/fetcher"
	"lottopredictor/internal/output"
)

/*
 * desc: AnalyzeWithDrawNumber()를 이용해 특정 회차 기준으로 3회 예측 수행
 * usage: drawNo 변수값 수정해서 테스트 진행
 */
func TestPredictionAndEvaluation(t *testing.T) {
	// 설정 로드
	config.LoadConfig("../config.json")

	// DB 연결 (경로 필요에 따라 조정)
	dbConn, err := db.InitDB("../database/lotto.db")
	if err != nil {
		t.Fatalf("DB 연결 실패: %v", err)
	} else {
		log.Printf("[DB] connection success\n")
	}

	drawNo := 1166 // 테스트 기준 회차

	// 3회 예측만 수행
	for i := 0; i < 3; i++ {
		analyzer.AnalyzeWithDrawNumber(dbConn, drawNo)
		log.Printf("[DB] AnalyzeWithDrawNumber(%d) success\n", drawNo)
	}

	// 다음 회차 실제 번호 불러오기
	actualData, err := fetcher.FetchDrawResult(drawNo + 1)
	if err != nil {
		t.Fatalf("당첨 번호 불러오기 실패: %v", err)
	}

	actual := []int{
		actualData.DrwtNo1, actualData.DrwtNo2, actualData.DrwtNo3,
		actualData.DrwtNo4, actualData.DrwtNo5, actualData.DrwtNo6,
	}
	bonus := actualData.BnusNo

	err = db.UpdatePredictionEvaluations(dbConn, actualData.DrwNo, actual, bonus)
	if err != nil {
		t.Fatalf("예측 결과 평가 실패: %v", err)
	}

	// 분석 결과 출력용: 마지막 예측 결과를 HTML + TXT로 저장
	result := analyzer.LoadLastPredictionResult(dbConn, drawNo+1)

	// 추천 결과 평가 정보도 가져와서 세팅
	rows, err := dbConn.Query(`
	SELECT percentage, rank 
	FROM prediction_results
	WHERE draw_number = ? ORDER BY meta_idx DESC, set_index
	LIMIT ?`, drawNo, len(result.SuggestionSets))
	if err == nil {
		var perc float64
		var rank int
		for rows.Next() {
			rows.Scan(&perc, &rank)
			result.Percentage = append(result.Percentage, perc)
			result.Ranks = append(result.Ranks, rank)
		}
		rows.Close()
	}

	// 결과 파일 저장
	outputPath := "result/lotto_analysis_%d"
	txtPath := fmt.Sprintf(outputPath+".txt", result.DrawNumber)
	htmlPath := fmt.Sprintf(outputPath+".html", result.DrawNumber)

	err = output.SaveAsTXT(result, txtPath)
	if err != nil {
		t.Logf("TXT 저장 실패: %v", err)
	}
	err = output.SaveAsHTML(result, htmlPath)
	if err != nil {
		t.Logf("HTML 저장 실패: %v", err)
	}

}
