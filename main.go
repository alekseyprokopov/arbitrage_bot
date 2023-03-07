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
	"time"
)

type Msg struct {
	Time    int64    `json:"time"`
	Channel string   `json:"channel"`
	Event   string   `json:"event"`
	Payload []string `json:"payload"`
	Auth    *Auth    `json:"auth"`
}

type Auth struct {
	Method string `json:"method"`
	KEY    string `json:"KEY"`
	SIGN   string `json:"SIGN"`
}

const (
	Key    = "YOUR_API_KEY"
	Secret = "YOUR_API_SECRETY"
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

type TickerData struct {
	base           string
	quote          string
	fee            float64
	minBaseAmount  float64
	minQuoteAmount float64
	askPrice       float64
	askValue       float64
	bidPrice       float64
	bidValue       float64
}

type ResponseData struct {
	Time    int    `json:"time"`
	TimeMs  int64  `json:"time_ms"`
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Result  struct {
		T  int64  `json:"t"`
		U  int    `json:"u"`
		S  string `json:"s"`
		B  string `json:"b"`
		B0 string `json:"B"`
		A  string `json:"a"`
		A0 string `json:"A"`
	} `json:"result"`
}

func main() {
	pairsUrl := "https://api.gateio.ws/api/v4/spot/currency_pairs"

	resp, err := DoRequest(pairsUrl, "")
	var responsePairs []Pair
	json.Unmarshal(*resp, &responsePairs)
	var pairs []string
	tickers := map[string]TickerData{}

	for _, item := range responsePairs {
		if item.TradeStatus != "untradable" {
			pairs = append(pairs, item.ID)

			fee, _ := strconv.ParseFloat(item.Fee, 64)
			minBaseAmount, _ := strconv.ParseFloat(item.MinBaseAmount, 64)
			mibQuoteAmount, _ := strconv.ParseFloat(item.MinQuoteAmount, 64)
			tickers[item.ID] = TickerData{
				base:           item.Base,
				quote:          item.Quote,
				fee:            fee,
				minQuoteAmount: mibQuoteAmount,
				minBaseAmount:  minBaseAmount,
			}
		}
	}
	fmt.Println("Получение пар завершено...")

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
			var response ResponseData
			json.Unmarshal(message, &response)
			askPrice, _ := strconv.ParseFloat(response.Result.A, 64)
			askValue, _ := strconv.ParseFloat(response.Result.A0, 64)
			bidPrice, _ := strconv.ParseFloat(response.Result.B, 64)
			bidValue, _ := strconv.ParseFloat(response.Result.B0, 64)

			pair := tickers[response.Result.S]
			pair.askPrice = askPrice
			pair.askValue = askValue
			pair.bidPrice = bidPrice
			pair.bidValue = bidValue
			tickers[response.Result.S] = pair
			fmt.Printf("recv: %+v\n", tickers[response.Result.S])
		}
	}()

	t := time.Now().Unix()

	// subscribe positions
	ordersMsg := NewMsg("spot.book_ticker", "subscribe", t, []string{"BTC_USDT"})
	err = ordersMsg.send(c)
	if err != nil {
		log.Fatal(err)

	}

	select {}
}

func DoRequest(urlAdd string, encodeQuery string) (*[]byte, error) {
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

type Pair struct {
	ID              string `json:"id"`
	Base            string `json:"base"`
	Quote           string `json:"quote"`
	Fee             string `json:"fee"`
	MinQuoteAmount  string `json:"min_quote_amount"`
	MinBaseAmount   string `json:"min_base_amount"`
	AmountPrecision int    `json:"amount_precision"`
	Precision       int    `json:"precision"`
	TradeStatus     string `json:"trade_status"`
	SellStart       int    `json:"sell_start"`
	BuyStart        int    `json:"buy_start"`
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
