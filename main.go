package main

import (
	"fmt"
	"github.com/gorilla/websocket"
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
				// 计算可用余额
				available := account.Balance / price //计算持仓价值
				// 更新账户余额和持仓
				account.Balance = 0          //余额归零
				account.Holding += available //持仓更新
				fmt.Printf("S value: %s,C value: %s,  ma value: %v\n", sValue, cValue, maValue)
				fmt.Printf("账户余额: %.2f, 持仓价值: %.2f, 盈亏: %.2f\n", account.Balance, account.Holding*price, account.Balance+account.Holding*price-1000)

			} else if price > maValue { //如果价格大于均值卖出
				fmt.Println("卖出信号")
				// 计算可用余额
				available := account.Holding * price //计算收益价值
				// 更新账户余额和持仓
				account.Holding = 0          //持仓更新
				account.Balance += available //余额更新
				fmt.Printf("S value: %s,C value: %s,  ma value: %v\n", sValue, cValue, maValue)
				fmt.Printf("账户余额: %.2f, 持仓价值: %.2f, 盈亏: %.2f\n", account.Balance, account.Holding*price, account.Balance+account.Holding*price-1000)

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
