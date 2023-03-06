package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	bybit := Platform{
		Name:   "bybit",
		ApiUrl: "https://api.bytick.com/v5/market/tickers",
	}

	//pairs := []string{"USDT", "USDC", "BUSD"}
	for {
		time.Sleep(time.Second)
		result, _ := bybit.spotData()
		fmt.Println(result)

	}
}

func (p *Platform) spotData() (*map[string]float64, error) {
	q := "category=spot"
	data, err := p.DoGetRequest(p.ApiUrl, q)
	if err != nil {
		return nil, fmt.Errorf("can't do getRequest to huobi API: %w", err)
	}
	var spotResponse SpotResponse
	if err := json.Unmarshal(*data, &spotResponse); err != nil {
		return nil, fmt.Errorf("can't unmarshall: %w", err)
	}

	result := map[string]float64{}
	//set := p.AllPairs
	list := spotResponse.Result.List

	//var dict = map[string]pair{}
	part1 := "USDT"
	for _, item1 := range list {
		ok1 := strings.Contains(item1.Symbol, part1)
		if ok1 {
			symbol1 := item1.Symbol
			price1, _ := strconv.ParseFloat(item1.Ask1Price, 64)
			//log.Printf("ASK: %s, BID: %s", item1.Ask1Price, item1.Bid1Price)
			part2 := strings.ReplaceAll(symbol1, part1, "")
			for _, item2 := range list {
				part3 := strings.ReplaceAll(item2.Symbol, part2, "")
				forwardPair := part2 + part3
				reversePair := part3 + part2

				if item2.Symbol == forwardPair {
					symbol2 := item2.Symbol
					price2, _ := strconv.ParseFloat(item2.Bid1Price, 64)

					for _, item3 := range list {
						ok3 := item3.Symbol == part3+part1 || item3.Symbol == part1+part3
						if ok3 {
							symbol3 := item3.Symbol
							price3, _ := strconv.ParseFloat(item3.Bid1Price, 64)

							profit := 100/price1*price2*price3 - 100

							if profit > 0 {
								fmt.Print("----forward---- \n")
								fmt.Printf("КРУГ: %s >>%s>>%s\n", symbol1, symbol2, symbol3)
								fmt.Printf("ЦЕНЫ: %f >>%f>>%f\n", price1, price2, price3)
								fmt.Printf("ПРОФИТ: %f \n", profit)
								fmt.Print("--------- \n")
							}

						}

					}
				} else if item2.Symbol == reversePair {
					symbol2 := item2.Symbol
					price2, _ := strconv.ParseFloat(item2.Ask1Price, 64)

					for _, item3 := range list {
						ok3 := item3.Symbol == part3+part1 || item3.Symbol == part1+part3

						if ok3 {
							symbol3 := item3.Symbol
							price3, _ := strconv.ParseFloat(item3.Bid1Price, 64)

							profit := 100/price1/price2*price3 - 100

							if profit > 0 {
								fmt.Print("----reverse----\n")
								fmt.Printf("КРУГ: %s >>%s>>%s\n", symbol1, symbol2, symbol3)
								fmt.Printf("ЦЕНЫ: %f >>%f>>%f\n", price1, price2, price3)
								fmt.Printf("ПРОФИТ: %f \n", profit)
								fmt.Print("---------\n")
							}

						}

					}
				}

			}

		}
	}
	return &result, err

}

type pair struct {
	Symbol   string
	askPrice float64
	bidPrice float64
	askSize  float64
	bidSrize float64
}

type SpotResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		Category string `json:"category"`
		List     []struct {
			Symbol                 string `json:"symbol"`
			LastPrice              string `json:"lastPrice"`
			IndexPrice             string `json:"indexPrice"`
			MarkPrice              string `json:"markPrice"`
			PrevPrice24H           string `json:"prevPrice24h"`
			Price24HPcnt           string `json:"price24hPcnt"`
			HighPrice24H           string `json:"highPrice24h"`
			LowPrice24H            string `json:"lowPrice24h"`
			PrevPrice1H            string `json:"prevPrice1h"`
			OpenInterest           string `json:"openInterest"`
			OpenInterestValue      string `json:"openInterestValue"`
			Turnover24H            string `json:"turnover24h"`
			Volume24H              string `json:"volume24h"`
			FundingRate            string `json:"fundingRate"`
			NextFundingTime        string `json:"nextFundingTime"`
			PredictedDeliveryPrice string `json:"predictedDeliveryPrice"`
			BasisRate              string `json:"basisRate"`
			DeliveryFeeRate        string `json:"deliveryFeeRate"`
			DeliveryTime           string `json:"deliveryTime"`
			Ask1Size               string `json:"ask1Size"`
			Bid1Price              string `json:"bid1Price"`
			Ask1Price              string `json:"ask1Price"`
			Bid1Size               string `json:"bid1Size"`
		} `json:"list"`
	} `json:"result"`
	RetExtInfo struct {
	} `json:"retExtInfo"`
	Time int64 `json:"time"`
}

type Platform struct {
	Name         string            `json:"platform_name"`
	Url          string            `json:"url"`
	ApiUrl       string            `json:"api_url"`
	Tokens       []string          `json:"platform_tokens"`
	TokensDict   map[string]string `json:"tokens_dict"`
	TradeTypes   []string          `json:"trade_types"`
	PayTypesDict map[string]string `json:"pay_types_dict"`
	AllPairs     map[string]bool   `json:"all_tokens"`
	Client       http.Client
}

func (p *Platform) DoGetRequest(urlAdd string, encodeQuery string) (*[]byte, error) {
	req, err := http.NewRequest(http.MethodGet, urlAdd, nil)
	if encodeQuery != "" {
		req.URL.RawQuery = encodeQuery
	}

	if err != nil {
		return nil, fmt.Errorf("can't do get request (%s): %w", p.Name, err)
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't get resposnse from DoGetRequest (%s): %w", p.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read info from response: %w", err)
	}

	return &body, err
}
