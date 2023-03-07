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
		Name:   "gate",
		ApiUrl: "https://api.gateio.ws/api/v4/spot/tickers",
	}

	//pairs := []string{"USDT", "USDC", "BUSD"}
	for {
		time.Sleep(time.Millisecond * 500)
		result, _ := bybit.spotData()
		fmt.Println(result)

	}
}

func (p *Platform) spotData() (*map[string]float64, error) {
	q := ""
	data, err := p.DoGetRequest(p.ApiUrl, q)
	if err != nil {
		return nil, fmt.Errorf("can't do getRequest to huobi API: %w", err)
	}

	var list []ListType
	if err := json.Unmarshal(*data, &list); err != nil {
		return nil, fmt.Errorf("can't unmarshall: %w", err)
	}

	result := map[string]float64{}
	//list := []ListType

	token := "USDT"
	for _, item1 := range list {
		item1 := item1
		if strings.Contains(item1.CurrencyPair, token) {
			for _, item2 := range list {
				item2 := item2
				go search(&list, token, &item1, &item2)
			}

		}
	}
	return &result, err

}

func search(list *[]ListType, token string, item1 *ListType, item2 *ListType) {
	part1 := token
	part2 := strings.ReplaceAll(item1.CurrencyPair, "_"+part1, "")

	part3 := strings.ReplaceAll(item2.CurrencyPair, "_", "")
	part3 = strings.ReplaceAll(part3, part2, "")

	var price2 float64
	var price3 float64

	forwardPair := part2 + "_" + part3
	reversePair := part3 + "_" + part2

	if item2.CurrencyPair == forwardPair {

		for _, item3 := range *list {
			item3 := item3
			go func() {
				ok3 := item3.CurrencyPair == part3+"_"+part1
				//|| item3.CurrencyPair == part1+"_"+part3
				if ok3 {
					price1, _ := strconv.ParseFloat(item1.LowestAsk, 64)
					price2, _ = strconv.ParseFloat(item2.HighestBid, 64)
					price3, _ = strconv.ParseFloat(item3.HighestBid, 64)

					profit := 100/price1*price2*price3 - 100

					if profit > 0 {
						fmt.Print("----forward---- \n")
						fmt.Printf("КРУГ: %s >>%s>>%s\n", item1.CurrencyPair, item2.CurrencyPair, item3.CurrencyPair)
						fmt.Printf("ЦЕНЫ: %f >>%f>>%f\n", price1, price2, price3)
						fmt.Printf("ПРОФИТ: %f \n", profit)
						fmt.Print("--------- \n")
					}

				}
			}()

		}

	} else if item2.CurrencyPair == reversePair {
		//log.Printf("part1: %s, part2: %s, part3: %s", part1, part2, part3)

		for _, item3 := range *list {
			item3 := item3
			go func() {
				ok3 := item3.CurrencyPair == part3+"_"+part1
				//|| item3.CurrencyPair == part1+"_"+part3
				if ok3 {
					price1, _ := strconv.ParseFloat(item1.LowestAsk, 64)
					price2, _ = strconv.ParseFloat(item2.LowestAsk, 64)
					price3, _ = strconv.ParseFloat(item3.HighestBid, 64)

					profit := 100/price1/price2*price3 - 100

					if profit > 0 {
						fmt.Print("---reverse----\n")
						fmt.Printf("КРУГ: %s >>%s>>%s\n", item1.CurrencyPair, item2.CurrencyPair, item3.CurrencyPair)
						fmt.Printf("ЦЕНЫ: %f >>%f>>%f\n", price1, price2, price3)
						fmt.Printf("ПРОФИТ: %f \n", profit)
						fmt.Print("---------\n")
					}

				}

			}()

		}
	}

}

func forwardPair(list []ListType, token string, item1 ListType, item2 ListType) {

}

type pair struct {
	Symbol   string
	askPrice float64
	bidPrice float64
	askSize  float64
	bidSrize float64
}

type ListType struct {
	CurrencyPair     string `json:"currency_pair"`
	Last             string `json:"last"`
	LowestAsk        string `json:"lowest_ask"`
	HighestBid       string `json:"highest_bid"`
	ChangePercentage string `json:"change_percentage"`
	ChangeUtc0       string `json:"change_utc0"`
	ChangeUtc8       string `json:"change_utc8"`
	BaseVolume       string `json:"base_volume"`
	QuoteVolume      string `json:"quote_volume"`
	High24H          string `json:"high_24h"`
	Low24H           string `json:"low_24h"`
	EtfNetValue      string `json:"etf_net_value"`
	EtfPreNetValue   string `json:"etf_pre_net_value"`
	EtfPreTimestamp  int    `json:"etf_pre_timestamp"`
	EtfLeverage      string `json:"etf_leverage"`
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
