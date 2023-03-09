package main

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"sync"
	"time"
)

func createStream(tickers sync.Map, pairs []string){
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
			//c.
			//fmt.Println(string(message))
			updateTickers(message, tickers)

		}
	}()

	t := time.Now().Unix()

	//subscribe positions
	ordersMsg := NewMsg("spot.book_ticker", "subscribe", t, pairs)

	err = ordersMsg.send(c)
	if err != nil {
		log.Fatal(err)
	}
}