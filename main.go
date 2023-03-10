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
	Key      = "3c437840e58b70a5a7d0cdb4763ef6a5"
	Secret   = "0607dc275ed118d4bbedc860bebfcf6f11f6c96bb254b5db304fc3d4a2fbc3cd"
	pairsUrl = "https://api.gateio.ws/api/v4/spot/currency_pairs"
)

var (
	count = 0
)

func main() {
	tickers, pairs, tokens := getPairs(pairsUrl)

	createStream(tickers, pairs, tokens)
	client := NewGateAPI(Key, Secret)

	for {
		findChains(&tickers, "USDT", client)
	}
	//time.Sleep(time.Second * 5)
	start := time.Now()

	//order := client.createOrder("TRX_USDT", "buy", "2", "")

	log.Println("BALANCE: ", client.getBalance())
	log.Println("Время ", time.Since(start))
	select {}

}

func findChains(tickers *sync.Map, token string, client *GateAPI) {
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
							forwardCheck(key1, key2, key3, tickers, client)
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

func forwardCheck(key1, key2, key3 string, tickers *sync.Map, client *GateAPI) {

	first, second, third := toChains(key1, key2, key3, tickers)
	profit := 100/first.askPrice*second.bidPrice*third.bidPrice - 100

	vol1USDT := first.askValue * first.askPrice
	vol2_1USDT := first.askPrice * second.bidValue
	vol2_2USDT := second.bidValue * second.bidPrice * third.bidPrice
	vol3USDT := third.bidValue * third.bidPrice

	minVol := MinOf(vol1USDT, vol2_1USDT, vol2_2USDT, vol3USDT)

	if profit > 0.5 && minVol > 1 && first.askPrice != 0 && second.askPrice != 0 && third.bidPrice != 0 {

		price1 := first.askPrice
		price2 := second.bidPrice
		price3 := third.bidPrice

		var initialAmount float64 = 5 //USDT
		amount1 := initialAmount / price1
		amount2 := amount1 / price2
		amount3 := amount2
		fmt.Println(price3, amount3)

		//client.createOrder("CAKE_USDTkey1", "buy", fmt.Sprint(initialAmount), "")
		//client.createOrder("CAKE_ETH", "buy", "0.0032", "1000")
		//client.createOrder("ETH_USDT", "buy", "0.0032", "1000")

		fmt.Println("forward", "key: ", key1, "key2: ", key2, "key3: ", key3)
		fmt.Println("first: ", first.askPrice, "second: ", second.bidPrice, "third: ", third.bidPrice)
		fmt.Println("firstV: ", first.askValue, "secondV: ", second.bidValue, "thirdV: ", third.bidValue)
		fmt.Println("firstminVusdt: ", vol1USDT, "secondVminUSDT1: ", vol2_1USDT, "secondVminUSDT2: ", vol2_2USDT, "thirdVminUSDT: ", vol3USDT)
		fmt.Println("minVOL: ", minVol)

		fmt.Println("profit: ", profit)

		//start := time.Now()
		//
		//log.Println("время запросов: ", time.Since(start))

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

func MinOf(vars ...float64) float64 {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
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
