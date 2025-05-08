// internal/analyzer/analyzer.go
package analyzer

import (
	"database/sql"
	"log"
	"lottopredictor/internal/common"
	"lottopredictor/internal/config"
	"math/rand"
	"sort"

	"lottopredictor/internal/db"
)

// PredictionResult 구조체는 분석 결과 + 추천 번호 세트를 포함한다.
type PredictionResult struct {
	DrawNumber     int
	Probabilities  map[int]float64
	Gaps           map[int]int
	TopFrequent    []int
	LeastFrequent  []int
	RecentMissing  []int
	FreqInLast10   []int
	SuggestionSets [][]int
	Percentage     []float64
	Ranks          []int
}

func Analyze(dbConn *sql.DB) *PredictionResult {
	rows, _ := dbConn.Query("SELECT draw_number, n1, n2, n3, n4, n5, n6 FROM lotto_results ORDER BY draw_number")
	defer rows.Close()

	totalDraws := 0
	latestDraw := 0
	draws := make(map[int][]int)
	count := make([]int, common.MaxLottoNum)
	lastSeen := make([]int, common.MaxLottoNum)
	last10freq := make([]int, common.MaxLottoNum)

	for rows.Next() {
		var drawNo int
		var nums [6]int
		rows.Scan(&drawNo, &nums[0], &nums[1], &nums[2], &nums[3], &nums[4], &nums[5])
		latestDraw = drawNo
		draws[drawNo] = nums[:]
		totalDraws++
		for _, n := range nums {
			count[n-1]++
			lastSeen[n-1] = drawNo
			if drawNo > latestDraw-common.Lookback {
				last10freq[n-1]++
			}
		}
	}

	probs := map[int]float64{} // probs: 등장 확률.	과거 통계 기반
	gaps := map[int]int{}      // gaps: 최근 미등장. 안 나온 번호 우선 반영
	for i := 0; i < common.MaxLottoNum; i++ {
		probs[i+1] = float64(count[i]) / float64(totalDraws) * 100
		gaps[i+1] = latestDraw - lastSeen[i]
	}

	db.SaveDrawProbabilities(dbConn, latestDraw, probs)
	db.SaveReappearanceProbabilities(dbConn, latestDraw, computeReappearance(draws, latestDraw))

	top10 := topNumbers(count, 10, true)
	least10 := topNumbers(count, 10, false)
	missing := []int{}
	for i := 0; i < common.MaxLottoNum; i++ {
		if latestDraw-lastSeen[i] >= config.AppConfig.GapThreshold {
			missing = append(missing, i+1)
		}
	}
	last10Top := topNumbers(last10freq, 10, true)

	suggestions := [][]int{}
	for i := 0; i < config.AppConfig.SuggestionSetCount; i++ {
		suggestions = append(suggestions, generateWeightedSample(probs, gaps, common.SetSize))
	}

	metaIdx, err := db.InsertPredictionMeta(dbConn, latestDraw+1)
	if err != nil {
		log.Fatal("메타 저장 실패:", err)
	}
	err = db.SavePredictionResults(dbConn, int64(latestDraw+1), metaIdx, suggestions)
	if err != nil {
		log.Fatal("추천 결과 저장 실패:", err)
	}

	return &PredictionResult{
		DrawNumber:     latestDraw + 1,
		Probabilities:  probs,
		Gaps:           gaps,
		TopFrequent:    top10,
		LeastFrequent:  least10,
		RecentMissing:  missing,
		FreqInLast10:   last10Top,
		SuggestionSets: suggestions,
	}
}

func topNumbers(arr []int, count int, descending bool) []int {
	type pair struct {
		Num  int
		Freq int
	}
	list := []pair{}
	for i, v := range arr {
		list = append(list, pair{i + 1, v})
	}
	sort.Slice(list, func(i, j int) bool {
		if descending {
			return list[i].Freq > list[j].Freq
		}
		return list[i].Freq < list[j].Freq
	})
	res := []int{}
	for i := 0; i < count && i < len(list); i++ {
		res = append(res, list[i].Num)
	}
	return res
}

func computeReappearance(draws map[int][]int, latest int) map[int]float64 {
	total := make([]int, common.MaxLottoNum)
	repeat := make([]int, common.MaxLottoNum)
	for i := 1; i < latest; i++ {
		curr := draws[i]
		next, ok := draws[i+1]
		if !ok {
			continue
		}
		check := map[int]bool{}
		for _, n := range curr {
			total[n-1]++
			check[n] = true
		}
		for _, n := range next {
			if check[n] {
				repeat[n-1]++
			}
		}
	}
	res := map[int]float64{}
	for i := 0; i < common.MaxLottoNum; i++ {
		if total[i] > 0 {
			res[i+1] = float64(repeat[i]) / float64(total[i]) * 100
		} else {
			res[i+1] = 0
		}
	}
	return res
}

func generateWeightedSample(probs map[int]float64, gaps map[int]int, count int) []int {
	selected := map[int]bool{}
	result := []int{}
	for len(result) < count {
		sum := 0.0
		weights := map[int]float64{}
		for i := 1; i <= common.MaxLottoNum; i++ {
			if selected[i] {
				continue
			}
			gap := gaps[i]
			// 시간 가중 평균 기반: 1 + (gap × multiplier)
			boost := 1.0 + float64(gap)*config.AppConfig.GAPBoostMultiplier
			w := probs[i] * boost
			weights[i] = w
			sum += w
		}
		r := randFloat() * sum
		acc := 0.0
		for i := 1; i <= common.MaxLottoNum; i++ {
			if selected[i] {
				continue
			}
			acc += weights[i]
			if r < acc {
				selected[i] = true
				result = append(result, i)
				break
			}
		}
	}
	sort.Ints(result)
	return result
}

func randFloat() float64 {
	return float64(rand.Intn(1000000)) / 1000000.0
}

// 1회부터 baseDraw 회차 직전까지의 확률을 구하고, 다음 회차를 예측
// 예측 결과를 prediction_results, prediction_meta 테이블에 저장하는 테스트/시뮬레이션용 분석 함수
func AnalyzeWithDrawNumber(dbConn *sql.DB, baseDraw int) *PredictionResult {
	targetDraw := baseDraw + 1
	log.Printf("[AnalyzeWithDrawNumber] 시작 - 기준 회차: %d → 예측 대상: %d\n", baseDraw, targetDraw)

	rows, _ := dbConn.Query(`
		SELECT draw_number, n1, n2, n3, n4, n5, n6 
		FROM lotto_results 
		WHERE draw_number <= ? ORDER BY draw_number`, baseDraw)
	defer rows.Close()

	totalDraws := 0
	draws := make(map[int][]int)
	count := make([]int, common.MaxLottoNum)
	lastSeen := make([]int, common.MaxLottoNum)
	last10freq := make([]int, common.MaxLottoNum)

	for rows.Next() {
		var drawNo int
		var nums [6]int
		rows.Scan(&drawNo, &nums[0], &nums[1], &nums[2], &nums[3], &nums[4], &nums[5])
		draws[drawNo] = nums[:]
		totalDraws++
		for _, n := range nums {
			count[n-1]++
			lastSeen[n-1] = drawNo
			if drawNo > baseDraw-config.AppConfig.LookbackRounds {
				last10freq[n-1]++
			}
		}
	}

	probs := map[int]float64{}
	gaps := map[int]int{}
	for i := 0; i < common.MaxLottoNum; i++ {
		probs[i+1] = float64(count[i]) / float64(totalDraws) * 100
		gaps[i+1] = targetDraw - lastSeen[i]
	}

	// 확률 저장은 baseDraw 기준
	db.SaveDrawProbabilities(dbConn, baseDraw, probs)
	db.SaveReappearanceProbabilities(dbConn, baseDraw, computeReappearance(draws, baseDraw))

	top10 := topNumbers(count, 10, true)
	least10 := topNumbers(count, 10, false)
	missing := []int{}
	for i := 0; i < common.MaxLottoNum; i++ {
		if targetDraw-lastSeen[i] >= config.AppConfig.GapThreshold {
			missing = append(missing, i+1)
		}
	}
	last10Top := topNumbers(last10freq, 10, true)

	suggestions := [][]int{}
	for i := 0; i < config.AppConfig.SuggestionSetCount; i++ {
		suggestions = append(suggestions, generateWeightedSample(probs, gaps, common.SetSize))
	}
	log.Printf("[AnalyzeWithDrawNumber] 추천 번호 생성 완료 (%d 세트), 저장 시작", len(suggestions))

	metaIdx, err := db.InsertPredictionMeta(dbConn, targetDraw)
	if err != nil {
		log.Fatal("메타 저장 실패:", err)
	} else {
		log.Printf("메타 저장 성공(drawNo:%d, metaIdx:%d)", targetDraw, metaIdx)
	}

	err = db.SavePredictionResults(dbConn, int64(targetDraw), metaIdx, suggestions)
	if err != nil {
		log.Fatal("추천 결과 저장 실패:", err)
	} else {
		log.Printf("추천 결과 저장 성공(drawNo:%d, metaIdx:%d)", targetDraw, metaIdx)
	}

	return &PredictionResult{
		DrawNumber:     targetDraw,
		Probabilities:  probs,
		Gaps:           gaps,
		TopFrequent:    top10,
		LeastFrequent:  least10,
		RecentMissing:  missing,
		FreqInLast10:   last10Top,
		SuggestionSets: suggestions,
	}
}

func LoadLastPredictionResult(dbConn *sql.DB, drawNo int) *PredictionResult {
	result := &PredictionResult{
		DrawNumber: drawNo,
	}

	// 가장 최근 meta_idx 가져오기
	row := dbConn.QueryRow(`
		SELECT MAX(meta_idx)
		FROM prediction_results
		WHERE draw_number = ?
	`, drawNo)

	var nullMetaIdx sql.NullInt64
	if err := row.Scan(&nullMetaIdx); err != nil {
		log.Printf("메타 인덱스 조회 실패: %v\n", err)
		return result
	}

	metaIdx := 0
	if nullMetaIdx.Valid {
		metaIdx = int(nullMetaIdx.Int64)
	} else {
		log.Printf("draw_number %d에 대한 meta_idx 결과 없음, 0으로 처리\n", drawNo)
		return result
	}

	// 해당 meta_idx의 추천 번호 가져오기
	rows, err := dbConn.Query(`
		SELECT num1, num2, num3, num4, num5, num6, percentage, rank
		FROM prediction_results
		WHERE draw_number = ? AND meta_idx = ?
		ORDER BY set_index ASC
	`, drawNo, metaIdx)
	if err != nil {
		log.Printf("추천번호 불러오기 실패: %v\n", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var n1, n2, n3, n4, n5, n6 int
		var perc float64
		var rank int
		err := rows.Scan(&n1, &n2, &n3, &n4, &n5, &n6, &perc, &rank)
		if err != nil {
			continue
		}
		result.SuggestionSets = append(result.SuggestionSets, []int{n1, n2, n3, n4, n5, n6})
		result.Percentage = append(result.Percentage, perc)
		result.Ranks = append(result.Ranks, rank)
	}

	return result
}
