package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const (

	pairsUrl = "https://api.gateio.ws/api/v4/spot/currency_pairs"
)

func main() {
	tickers, pairs := getPairs(pairsUrl)
	createStream(tickers, pairs)

	for {
		findChains(&tickers, "USDT")
	}
	select {}

}

func findChains(tickers *sync.Map, token string) {
	start := time.Now()
	tickers.Range(func(symbol, tickerData any) bool {
		key1 := symbol.(string)
		value1 := tickerData.(TickerData)
		if value1.quote == token {
			part := value1.base

			tickers.Range(func(symbol, tickerData any) bool {
				key2 := symbol.(string)
				value2 := tickerData.(TickerData)

				if part == value2.base {
					tickers.Range(func(symbol, tickerData any) bool {
						key3 := symbol.(string)
						value3 := tickerData.(TickerData)
						if value2.quote == value3.base && value3.quote == token {
							forwardCheck(key1, key2, key3, tickers)
						}
						return true
					})

				}
				if part == value2.quote {
					tickers.Range(func(symbol, tickerData any) bool {
						key3 := symbol.(string)
						value3 := tickerData.(TickerData)
						if value2.base == value3.base && value3.quote == token {
							reverseCheck(key1, key2, key3, tickers)

						}
						return true
					})
				}
				return true
			})

		}
		return true
	})

	log.Println(time.Since(start))
}

func forwardCheck(key1, key2, key3 string, tickers *sync.Map) {

	first, second, third := toChains(key1, key2, key3, tickers)
	profit := 100/first.askPrice/second.askPrice/second.bidPrice - 100

	if profit > 1 && first.askPrice != 0 && second.askPrice != 0 && third.bidPrice != 0 {
		//fmt.Println("forward", "key: ", key1, "key2: ", key2, "key3: ", key3)
		//fmt.Println("first: ", first.askPrice, "second: ", second.bidPrice, "third: ", third.bidPrice)
		//fmt.Println("firstV: ", first.askValue, "secondV: ", second.bidValue, "thirdV: ", third.bidValue)
		//fmt.Println("firstminVusdt: ", first.askValue*first.askPrice, "secondVminUSDT1: ", first.askPrice*second.bidValue, "secondVminUSDT2: ", second.bidValue*second.bidPrice*third.bidPrice, "thirdVminUSDT: ", third.bidValue*third.bidPrice)
		//
		//fmt.Println("profit: ", profit)
		start := time.Now()
		doRequest("https://api.gateio.ws/api/v4/spot/orders", "")
		doRequest("https://api.gateio.ws/api/v4/spot/orders", "")
		doRequest("https://api.gateio.ws/api/v4/spot/orders", "")
		log.Println("время запросов: ", time.Since(start))

	}

}
func reverseCheck(key1, key2, key3 string, tickers *sync.Map) {
	fmt.Sprint("reverse", "key: ", key1, "key2: ", key2, "key3: ", key3)
}

func toChains(key1, key2, key3 string, tickers *sync.Map) (TickerData, TickerData, TickerData) {
	value1, _ := tickers.Load(key1)
	value2, _ := tickers.Load(key2)
	value3, _ := tickers.Load(key3)

	return value1.(TickerData), value2.(TickerData), value3.(TickerData)
}
func doRequest(urlAdd string, encodeQuery string) (*[]byte, error) {
	Client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlAdd, nil)
	if encodeQuery != "" {
		req.URL.RawQuery = encodeQuery
	}

	if err != nil {
		return nil, fmt.Errorf("can't do get request (%s): %w", "gate", err)
	}

	resp, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't get resposnse from DoGetRequest (%s): %w", "gate", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read info from response: %w", err)
	}

	return &body, err
}

func createOrder() {

}

func doPostRequest() {

}

//pingMsg := NewMsg("spot.ping", "", t, []string{})
//err = pingMsg.send(c)
//if err != nil {
//	log.Fatal(err)
//}

//spot tickers
//spotTickers := NewMsg("spot.tickers", "subscribe", t, []string{"BTC_USDT"})
//err = spotTickers.send(c)
//if err != nil {
//	log.Fatal(err)
//
//}
