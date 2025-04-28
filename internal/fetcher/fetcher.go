// internal/fetcher/fetcher.go
package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DrawData struct {
	ReturnValue string `json:"returnValue"`
	DrwNo       int    `json:"drwNo"`
	DrwtNo1     int    `json:"drwtNo1"`
	DrwtNo2     int    `json:"drwtNo2"`
	DrwtNo3     int    `json:"drwtNo3"`
	DrwtNo4     int    `json:"drwtNo4"`
	DrwtNo5     int    `json:"drwtNo5"`
	DrwtNo6     int    `json:"drwtNo6"`
	BnusNo      int    `json:"bnusNo"`
	DrwNoDate   string `json:"drwNoDate"`
}

const apiURL = "https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d"

func FetchDrawData(drawNo int) (*DrawData, error) {
	url := fmt.Sprintf(apiURL, drawNo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data DrawData
	err = json.Unmarshal(body, &data)
	if err != nil || data.DrwNo == 0 {
		return nil, fmt.Errorf("no data or invalid response for draw %d", drawNo)
	}
	return &data, nil
}

func FetchDrawResult(drawNo int) (*DrawData, error) {
	url := fmt.Sprintf("https://www.dhlottery.co.kr/common.do?method=getLottoNumber&drwNo=%d", drawNo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var data DrawData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if data.ReturnValue != "success" {
		return nil, fmt.Errorf("회차 %d 결과 없음", drawNo)
	}

	return &data, nil
}
