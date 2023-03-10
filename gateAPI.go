package main

import (
	"context"
	"github.com/gateio/gateapi-go/v6"
	"log"
)

type GateAPI struct {
	*gateapi.APIClient
	key    string
	secret string
}

func NewGateAPI(key, secret string) *GateAPI {
	conf := gateapi.NewConfiguration()
	conf.Key = key
	conf.Secret = secret

	client := gateapi.NewAPIClient(conf)
	return &GateAPI{
		client,
		key,
		secret,
	}
}

func (client *GateAPI) getBalance() gateapi.TotalBalance {
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    client.key,
		Secret: client.secret,
	})
	balance, _, _ := client.WalletApi.GetTotalBalance(ctx, nil)
	return balance
}

func (client *GateAPI) createOrder(pair, side, amount, price string) *gateapi.Order {
	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    client.key,
		Secret: client.secret,
	})

	newOrder := gateapi.Order{
		CurrencyPair: pair,
		Type:         "market",
		Account:      "spot",
		Side:         side,
		Amount:       amount,
		//Price:        price,//not for market
		TimeInForce: "ioc", //not for market
		AutoBorrow:  false,
	}

	createdOrder, response, err := client.SpotApi.CreateOrder(ctx, newOrder)
	if err != nil {
		log.Fatal("can't create order: ", err)
	}
	log.Printf("RESPONSE: ", response)

	log.Printf("order created with ID: %s, status: %s\n", createdOrder.Id, createdOrder.Status)
	return &createdOrder
}
