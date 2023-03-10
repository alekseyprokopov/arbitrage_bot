package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func sign(channel, event string, t int64) string {
	message := fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, t)
	h2 := hmac.New(sha512.New, []byte(Secret))
	io.WriteString(h2, message)
	return hex.EncodeToString(h2.Sum(nil))
}

func (msg *Msg) sign() {
	signStr := sign(msg.Channel, msg.Event, msg.Time)
	msg.Auth = &Auth{
		Method: "api_key",
		KEY:    Key,
		SIGN:   signStr,
	}
}

func createStream(tickers sync.Map, pairs []string, tokens sync.Map) {
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
			a := struct {
				Channel string `json:"channel"`
			}{}
			json.Unmarshal(message, &a)

			switch a.Channel {
			case "spot.book_ticker":
				updateTickers(message, tickers)
			case "spot.balances":
				updateBalance(message, tokens)
				//case "spot.orders":
				//	log.Println(string(message))
			}

		}
	}()

	t := time.Now().Unix()

	//subscribe positions
	ordersMsg := NewMsg("spot.book_ticker", "subscribe", t, pairs)
	err = ordersMsg.send(c)
	if err != nil {
		log.Fatal(err)
	}

	balanceMsg := NewMsg("spot.balances", "subscribe", t, nil)
	balanceMsg.sign()
	err = balanceMsg.send(c)
	if err != nil {
		log.Fatal(err)
	}

	//spotOrdersMsg := NewMsg("spot.orders", "subscribe", t, []string{"!all"})
	//spotOrdersMsg.sign()
	//err = spotOrdersMsg.send(c)
	//if err != nil {
	//	log.Fatal(err)
	//}
}

func updateTickers(message []byte, tickers sync.Map) {
	var response tickersUpdateData
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

func updateBalance(message []byte, tokens sync.Map) {
	var response balanceUpdateData
	json.Unmarshal(message, &response)
	for _, item := range response.Result {
		_, ok := tokens.Load(item.Currency)
		if ok {
			tokens.Store(item.Currency, item.Available)
		}
	}
}
