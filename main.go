package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	Key      = "YOUR_API_KEY"
	Secret   = "YOUR_API_SECRETY"
	pairsUrl = "https://api.gateio.ws/api/v4/spot/currency_pairs"
)

//func sign(channel, event string, t int64) string {
//	message := fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, t)
//	h2 := hmac.New(sha512.New, []byte(Secret))
//	io.WriteString(h2, message)
//	return hex.EncodeToString(h2.Sum(nil))
//}
//
//func (msg *Msg) sign() {
//	signStr := sign(msg.Channel, msg.Event, msg.Time)
//	msg.Auth = &Auth{
//		Method: "api_key",
//		KEY:    Key,
//		SIGN:   signStr,
//	}
//}

func (msg *Msg) send(c *websocket.Conn) error {
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msgByte)
}

func NewMsg(channel, event string, t int64, payload []string) *Msg {
	return &Msg{
		Time:    t,
		Channel: channel,
		Event:   event,
		Payload: payload,
	}
}

func main() {
	tickers, pairs := getPairs(pairsUrl)
	u := url.URL{Scheme: "wss", Host: "api.gateio.ws", Path: "/ws/v4/"}
	websocket.DefaultDialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)

	}
	c.SetPingHandler(nil)

	// read msg
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				c.Close()
				log.Fatal(err)
			}
			updateTickers(message, tickers)

		}
	}()

	t := time.Now().Unix()

	// subscribe positions
	ordersMsg := NewMsg("spot.book_ticker", "subscribe", t, pairs)
	err = ordersMsg.send(c)
	if err != nil {
		log.Fatal(err)

	}
	for {
		spotData(&tickers, "USDT")
	}
	select {}

}

func spotData(tickers *sync.Map, token string) {
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

	first, second, third := getChains(key1, key2, key3, tickers)
	profit := 100/first.askPrice/second.askPrice/second.bidPrice - 100
	firstVol


	if profit > 1 && first.askPrice != 0 && second.askPrice != 0 && third.bidPrice != 0 {
		fmt.Println("forward", "key: ", key1, "key2: ", key2, "key3: ", key3)
		fmt.Println("first: ", first.askPrice, "second: ", second.bidPrice, "third: ", third.bidPrice)
		fmt.Println("firstV: ", first.askValue, "secondV: ", second.bidValue, "thirdV: ", third.bidValue)
		fmt.Println("firstminVusdt: ", first.askValue*first.askPrice, "secondVminUSDT1: ", first.askPrice*second.bidValue, "secondVminUSDT2: ", second.bidValue*second.bidPrice*third.bidPrice, "thirdVminUSDT: ", third.bidValue*third.bidPrice)

		fmt.Println("profit: ", profit)
	}

}
func reverseCheck(key1, key2, key3 string, tickers *sync.Map) {
	fmt.Sprint("reverse", "key: ", key1, "key2: ", key2, "key3: ", key3)
}

func getChains(key1, key2, key3 string, tickers *sync.Map) (TickerData, TickerData, TickerData) {
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

func getPairs(pairsUrl string) (tickers sync.Map, pairs []string) {
	resp, err := doRequest(pairsUrl, "")
	if err != nil {
		log.Fatalf("can't get pairs: %w", err)
	}
	var responsePairs []Pair
	json.Unmarshal(*resp, &responsePairs)

	tickers = sync.Map{}

	for _, item := range responsePairs {
		if item.TradeStatus != "untradable" {
			pairs = append(pairs, item.ID)

			fee, _ := strconv.ParseFloat(item.Fee, 64)
			minBaseAmount, _ := strconv.ParseFloat(item.MinBaseAmount, 64)
			mibQuoteAmount, _ := strconv.ParseFloat(item.MinQuoteAmount, 64)
			result := TickerData{
				base:           item.Base,
				quote:          item.Quote,
				fee:            fee,
				minQuoteAmount: mibQuoteAmount,
				minBaseAmount:  minBaseAmount,
			}
			tickers.Store(item.ID, result)
		}
	}
	fmt.Println("Получение пар завершено...")
	return tickers, pairs
}

func updateTickers(message []byte, tickers sync.Map) {
	var response ResponseData
	json.Unmarshal(message, &response)
	askPrice, _ := strconv.ParseFloat(response.Result.A, 64)
	askValue, _ := strconv.ParseFloat(response.Result.A0, 64)
	bidPrice, _ := strconv.ParseFloat(response.Result.B, 64)
	bidValue, _ := strconv.ParseFloat(response.Result.B0, 64)

	pair, ok := tickers.Load(response.Result.S)
	if ok {
		pair := pair.(TickerData)
		pair.askPrice = askPrice
		pair.askValue = askValue
		pair.bidPrice = bidPrice
		pair.bidValue = bidValue
		tickers.Store(response.Result.S, pair)
	}

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
