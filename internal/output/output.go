// internal/output/output.go
package output

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"lottopredictor/internal/analyzer"
	"lottopredictor/internal/common"
)

func SaveAsTXT(result *analyzer.PredictionResult, path string) error {
	os.MkdirAll("result", os.ModePerm)
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("회차: %d\n\n", result.DrawNumber))

	builder.WriteString("[상위 10 확률 번호]\n")
	top := topSorted(result.Probabilities, true)
	for i := 0; i < 10 && i < len(top); i++ {
		num := top[i]
		builder.WriteString(fmt.Sprintf("%2d: %6.3f%% (간격: %d)\n", num, result.Probabilities[num], result.Gaps[num]))
	}

	builder.WriteString("\n[추천 번호 세트]\n")
	for i, set := range result.SuggestionSets {
		line := fmt.Sprintf("추천 %2d: %v", i+1, set)
		// 평가 정보가 있다면 추가
		if i < len(result.Percentage) && i < len(result.Ranks) {
			line += fmt.Sprintf("  | 일치율: %5.1f%%, 등수: %d", result.Percentage[i], result.Ranks[i])
		}
		builder.WriteString(line + "\n")
	}

	builder.WriteString("\n[최근 10회 미등장 번호]\n")
	builder.WriteString(fmt.Sprintf("%v\n", result.RecentMissing))

	builder.WriteString("\n[최근 10회 출현 빈도 높은 번호]\n")
	for _, num := range result.FreqInLast10 {
		builder.WriteString(fmt.Sprintf("%2d: %6.3f%%\n", num, result.Probabilities[num]))
	}

	builder.WriteString("\n[가장 많이 등장한 번호 Top 10]\n")
	for _, num := range result.TopFrequent {
		builder.WriteString(fmt.Sprintf("%2d: %6.3f%%\n", num, result.Probabilities[num]))
	}

	builder.WriteString("\n[가장 적게 등장한 번호 Top 10]\n")
	for _, num := range result.LeastFrequent {
		builder.WriteString(fmt.Sprintf("%2d: %6.3f%%\n", num, result.Probabilities[num]))
	}

	os.WriteFile(path, []byte(builder.String()), 0644)

	return nil
}

func SaveAsHTML(result *analyzer.PredictionResult, path string) error {
	os.MkdirAll("result", os.ModePerm)
	html := strings.Builder{}
	html.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8">
	<title>Lotto 분석 결과</title>
	<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
	</head><body><h1>회차 ` + fmt.Sprint(result.DrawNumber) + ` 분석</h1>
	<h2>상위 10 확률 번호</h2><ul>`)

	top := topSorted(result.Probabilities, true)
	for i := 0; i < 10 && i < len(top); i++ {
		n := top[i]
		html.WriteString(fmt.Sprintf("<li>%2d: %.3f%% (간격 %d)</li>", n, result.Probabilities[n], result.Gaps[n]))
	}
	html.WriteString(`</ul><h2>추천 번호 세트</h2><ul>`)
	for _, set := range result.SuggestionSets {
		html.WriteString("<li>")
		for i, n := range set {
			html.WriteString(fmt.Sprintf("%d", n))
			if i < len(set)-1 {
				html.WriteString(", ")
			}
		}
		html.WriteString("</li>")
	}

	html.WriteString(`<h2>최근 10회 출현 빈도 높은 번호</h2><table border="1"><tr><th>번호</th><th>확률 (%)</th></tr>`)
	for _, num := range result.FreqInLast10 {
		html.WriteString(fmt.Sprintf("<tr><td>%2d</td><td>%.3f</td></tr>", num, result.Probabilities[num]))
	}
	html.WriteString("</table>")

	html.WriteString(`<h2>가장 많이 등장한 번호 Top 10</h2><table border="1"><tr><th>번호</th><th>확률 (%)</th></tr>`)
	for _, num := range result.TopFrequent {
		html.WriteString(fmt.Sprintf("<tr><td>%2d</td><td>%.3f</td></tr>", num, result.Probabilities[num]))
	}
	html.WriteString("</table>")

	html.WriteString(`<h2>가장 적게 등장한 번호 Top 10</h2><table border="1"><tr><th>번호</th><th>확률 (%)</th></tr>`)
	for _, num := range result.LeastFrequent {
		html.WriteString(fmt.Sprintf("<tr><td>%2d</td><td>%.3f</td></tr>", num, result.Probabilities[num]))
	}
	html.WriteString("</table>")

	html.WriteString(`</ul><canvas id="chart" width="900" height="400"></canvas>
	<script>
	const ctx = document.getElementById('chart').getContext('2d');
	new Chart(ctx, {
		type: 'bar',
		data: {
			labels: [`)
	for i := 1; i <= common.MaxLottoNum; i++ {
		html.WriteString(fmt.Sprintf("'%d'", i))
		if i != common.MaxLottoNum {
			html.WriteString(",")
		}
	}
	html.WriteString(`],
			datasets: [{
				label: '등장 확률 (%)',
				data: [`)
	for i := 1; i <= common.MaxLottoNum; i++ {
		html.WriteString(fmt.Sprintf("%.3f", result.Probabilities[i]))
		if i != common.MaxLottoNum {
			html.WriteString(",")
		}
	}
	html.WriteString(`],
				borderWidth: 1,
				backgroundColor: 'rgba(75,192,192,0.6)'
			}]
		},
		options: {
			scales: {
				y: {
					beginAtZero: true,
					ticks: {
						callback: function(value) {
							return value + '%';
						}
					},
					grid: {
						drawTicks: true
					}
				}
			},
			plugins: {
				annotation: {
					annotations: [{
						type: 'line',
						yMin: 13.04,
						yMax: 13.04,
						borderColor: 'red',
						borderWidth: 2,
						label: {
							enabled: true,
							content: '평균선(13%)'
						}
					}]
				}
			}
		}
	});
	</script></body></html>`)

	html.WriteString(`<h2>추천 결과 평가</h2>
	<table border="1" cellpadding="8" cellspacing="0">
	<tr>
	  <th>세트</th>
	  <th>추천 번호</th>
	  <th>일치율 (%)</th>
	  <th>등수</th>
	</tr>
	`)

	for i, set := range result.SuggestionSets {
		html.WriteString(fmt.Sprintf("<tr><td>%d</td><td>", i+1))
		for j, n := range set {
			html.WriteString(fmt.Sprintf("%2d", n))
			if j < len(set)-1 {
				html.WriteString(", ")
			}
		}
		html.WriteString("</td>")

		// 평가 정보가 있으면 출력, 없으면 빈칸
		percent := ""
		rank := ""
		if i < len(result.Percentage) {
			percent = fmt.Sprintf("%.1f", result.Percentage[i])
		}
		if i < len(result.Ranks) {
			rank = fmt.Sprintf("%d", result.Ranks[i])
		}
		html.WriteString(fmt.Sprintf("<td>%s</td><td>%s</td></tr>\n", percent, rank))
	}
	html.WriteString("</table><br>`")

	os.WriteFile(path, []byte(html.String()), 0644)

	return nil
}

func topSorted(m map[int]float64, desc bool) []int {
	type kv struct {
		Key int
		Val float64
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		if desc {
			return ss[i].Val > ss[j].Val
		}
		return ss[i].Val < ss[j].Val
	})
	res := make([]int, len(ss))
	for i, kv := range ss {
		res[i] = kv.Key
	}
	return res
}
