package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Get("https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("BTC-USDT price: %s\n", resp.Body())
}
