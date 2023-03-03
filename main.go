package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/parnurzeal/gorequest"
	"github.com/shopspring/decimal"
	"log"
	"strconv"
	"strings"
	"time"
)

var url = "https://api4.binance.com"
var key = "yCij8LdqY0u7boPzB4GN9ACzMAaqaS4L1ezUPwGCUHaQmUqjYQxc6mZM6xm1WQDw"
var secret = "7TuigXuGSdRn6URlefpyVC71M8vFf1ybxcOWQBnQkVOUsXdiPlYqFZZrfD0QRVou"

type AccountAmount struct {
	Balances []Amount
}

// Amount 账户
type Amount struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// 移动平均线
type MovingAverage struct {
	Period int
	Data   []float64
}

func (ma *MovingAverage) Add(val float64) {
	if len(ma.Data) < ma.Period {
		ma.Data = append(ma.Data, val)
		fmt.Println(val)
		return
	}
	// 移除队头数据
	ma.Data = append(ma.Data[1:], val)
}

func (ma *MovingAverage) Value() float64 {
	if len(ma.Data) < ma.Period {
		return 0
	}
	sum := 0.0
	for _, val := range ma.Data {
		sum += val
	}
	return sum / float64(ma.Period)
}

func QT() {
	// 建立 WebSocket 连接
	wsURL := "wss://stream.binancefuture.com/ws/btcusdt@ticker"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()
	//初始化移动平均线
	ma := MovingAverage{Period: 50}
	pos := 0
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Fatal("read error:", err)
		}
		jsonStr := string(message)
		//获取s字段位置
		sIndex := strings.Index(jsonStr, `"s":`)
		if sIndex == -1 {
			fmt.Println("C field not found")
			return
		}
		endSIndex := sIndex + strings.Index(jsonStr[sIndex:], ",")
		if endSIndex == -1 {
			fmt.Println("Error: unable to extract C field value")
			return
		}
		//sValue := jsonStr[sIndex+4 : endSIndex]
		//获取C字段位置
		CIndex := strings.Index(jsonStr, `"c":`)
		if CIndex == -1 {
			fmt.Println("C field not found")
			return
		}
		endCIndex := CIndex + strings.Index(jsonStr[CIndex:], ",")
		if endCIndex == -1 {
			fmt.Println("Error: unable to extract C field value")
			return
		}
		cValue := strings.Trim(jsonStr[CIndex+4:endCIndex], "\"")
		fmt.Println(cValue)
		// 将C值转换为float64类型
		price, err := strconv.ParseFloat(cValue, 64)
		//fmt.Println(cValue, "%f", &price)
		// 添加到移动平均线中
		ma.Add(price)
		// 获取移动平均线值
		min := decimal.NewFromFloat(0.0001)
		maValue := ma.Value()
		if maValue != 0 {
			if price > maValue { //如果小于均值买入
				pos += 1
				if pos == 5 {
					pos = 0
					fmt.Println("买入信号")
					usdt := GetUsdt()
					if usdt.LessThanOrEqual(decimal.NewFromInt(10)) {
						fmt.Println("usdt 额度不足")
						continue
					}
					err := Order("BUY", usdt.Div(decimal.NewFromFloat(price)).Sub(min).StringFixedBank(5))
					if err != nil {
						fmt.Println(err)
						continue
					}
					fmt.Println("买入成功:" + usdt.String())
				}
			} else if price < maValue { //如果价格大于均值卖出
				pos -= 1
				if pos == -5 {
					pos = 0
					fmt.Println("卖出信号")
					btc := GetBtc()
					if btc.LessThanOrEqual(decimal.NewFromFloat(0.0001)) {
						fmt.Println("btc 额度不足")
						continue
					}
					err := Order("SELL", btc.Sub(decimal.NewFromFloat(0.00001)).StringFixed(5))
					if err != nil {
						fmt.Println(err)
						continue

					}
					fmt.Println("卖出成功:" + btc.String())
				}
			}
		}
	}
}

func main() {
	QT()
}

func GetUsdt() decimal.Decimal {
	balance := GetBalance()
	for _, v := range balance {
		if v.Asset == "USDT" {
			usdt, _ := decimal.NewFromString(v.Free)
			return usdt
		}
	}
	return decimal.Zero
}

func GetBtc() decimal.Decimal {
	balance := GetBalance()
	for _, v := range balance {
		if v.Asset == "BTC" {
			usdt, _ := decimal.NewFromString(v.Free)
			return usdt
		}
	}
	return decimal.Zero
}

func GetBalance() []Amount {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	signature := hmacSha256("timestamp="+timestamp, secret)
	rsp := AccountAmount{}
	_, _, errs := gorequest.New().Get(url+"/api/v3/account?timestamp="+timestamp+"&signature="+signature).AppendHeader("X-MBX-APIKEY", key).EndStruct(&rsp)
	if len(errs) != 0 {
		fmt.Println(errs)
		return nil
	}
	return rsp.Balances
}

// Order
func Order(side, amount string) error {
	parameter := fmt.Sprintf("quantity=%s&side=%s&symbol=BTCUSDT&timestamp=%d&type=MARKET", amount, side, time.Now().UnixMilli())
	signature := hmacSha256(parameter, secret)
	parameter = parameter + "&signature=" + signature
	_, s, errs := gorequest.New().Post(url+"/api/v3/order"). //+parameter+"&signature="+signature
		SendString(parameter).
		AppendHeader("X-MBX-APIKEY", key).End()
	if len(errs) != 0 {
		fmt.Println(errs)
		return errs[0]
	}
	fmt.Println(s)
	return nil
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
