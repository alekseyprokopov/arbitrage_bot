package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
)



func getPairs(pairsUrl string) (tickers sync.Map, pairs []string, tokens sync.Map) {
	resp, err := doRequest(pairsUrl, "")
	if err != nil {
		log.Fatalf("can't get pairs: %w", err)
	}
	var responsePairs []Pair
	json.Unmarshal(*resp, &responsePairs)

	tickers = sync.Map{}
	tokens = sync.Map{}

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
			tokens.Store(item.Base, "0")
			tokens.Store(item.Quote, "0")
		}
	}
	fmt.Println("Получение пар завершено...")
	return tickers, pairs, tokens
}

