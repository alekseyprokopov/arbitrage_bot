package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

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