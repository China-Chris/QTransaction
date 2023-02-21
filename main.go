package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"log"
	"strconv"
	"strings"
)

type Account struct {
	Balance float64 // 账户余额
	Holding float64 // 持仓
}

type Candle struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func init() {
	//config.ParseConfig("./configs/cfg.yml")
	//logger.InitLog()
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

func btc() {
	var (
		apiKey    = "G30kZdFVnK2zugrWmw91lZdqDWzjsiLHY7eAm10IWaV2rc5Uq6LE7eBSE2J9NKK9Copy"
		secretKey = "OVNWYgvbVeQ6D7p1iCyQV3AvFVgolLiOacFspQikcSAl211i1JoMqTe2rt6aoqZg"
	)
	//创建客户端
	client := binance.NewClient(apiKey, secretKey)

	// 定义交易货币对和交易参数
	//symbol := "BTCUSDT"
	//tradePercent := 0.1 // 每次交易使用不超过10%的资金
	//futuresClient := binance.NewFuturesClient(apiKey, secretKey)   // USDT-M Futures
	//deliveryClient := binance.NewDeliveryClient(apiKey, secretKey) // Coin-M Futures

	// 建立 WebSocket 连接
	wsURL := "wss://stream.binancefuture.com/ws/btcusdt@ticker"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()

	//初始化移动平均线
	ma := MovingAverage{Period: 50}
	// 初始化账户
	account := Account{Balance: 1000}
	// 接收 WebSocket 数据
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
		sValue := jsonStr[sIndex+4 : endSIndex]
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
		maValue := ma.Value()
		if maValue != 0 {
			if price < maValue { //如果小于均值买入
				fmt.Println("买入信号")
				//计算买入数量
				availableMr := account.Balance
				quantityMr := decimal.NewFromFloat(availableMr).Div(decimal.NewFromFloat(price)).Floor()
				//执行买入操作
				order, err := client.NewCreateOrderService().Symbol("BTCUSDT").
					Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
					TimeInForce(binance.TimeInForceTypeGTC).Quantity(quantityMr.String()).
					Price(fmt.Sprintf("%.8f", price)).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(order)
				// 计算可用余额
				available := account.Balance / price //计算持仓价值
				// 更新账户余额和持仓
				account.Balance = 0          //余额归零
				account.Holding += available //持仓更新
				fmt.Printf("S value: %s,C value: %s,  ma value: %v\n", sValue, cValue, maValue)
				fmt.Printf("账户余额: %.2f, 持仓价值: %.2f, 盈亏: %.2f\n", account.Balance, account.Holding*price, account.Balance+account.Holding*price-10)

			} else if price > maValue { //如果价格大于均值卖出
				if account.Holding != 0 {
					fmt.Println("卖出信号")
					// 调用API获取账户信息
					accountBNB, err := client.NewGetAccountService().Do(context.Background())
					if err != nil {
						fmt.Println(err)
						return
					}
					// 获取账户中持有的BTC数量
					var holdingBtc decimal.Decimal
					for _, b := range accountBNB.Balances {
						if b.Asset == "BTC" {
							holdingBtc, _ = decimal.NewFromString(b.Free)
							break
						}
					}
					quantityDec := holdingBtc.Mul(decimal.NewFromFloat(1)) //
					quantity, _ := quantityDec.Round(6).Float64()          // 转换为float64类型，保留6位小数

					order, err := client.NewCreateOrderService().Symbol("BTCUSDT").
						Side(binance.SideTypeSell).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(decimal.NewFromFloat(quantity).String()).
						Price(fmt.Sprintf("%.8f", price)).Do(context.Background())
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println(order)
					// 计算可用余额
					available := account.Holding * price //计算收益价值
					// 更新账户余额和持仓
					account.Holding = 0          //持仓更新
					account.Balance += available //余额更新
					fmt.Printf("S value: %s,C value: %s,  ma value: %v\n", sValue, cValue, maValue)
					fmt.Printf("账户余额: %.2f, 持仓价值: %.2f, 盈亏: %.2f\n", account.Balance, account.Holding*price, account.Balance+account.Holding*price-10)
				}
			}
		}
	}
}

//func floatToString(f float64) string {
//	return strconv.FormatFloat(f, 'f', -1, 64)
//}

func main() {
	btc()
}
