package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
)

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
