package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	binance "github.com/adshao/go-binance"
	"github.com/go-redis/redis/v7"
)

// BinanceWatcher 币安 websocket 数据监听服务
type BinanceWatcher struct {
	redisClient *redis.Client
	config      *BinanceConf
	depthDoneC  chan struct{}
	depthStopC  chan struct{}
	tickers     []*time.Ticker
}

// Init 初始化依赖注入
func (bw *BinanceWatcher) Init(client *redis.Client, config *BinanceConf) {
	bw.redisClient = client
	bw.config = config
}

// WatchDepth 监听盘口价格
func (bw *BinanceWatcher) WatchDepth() {
	wsPartialDepthHandler := func(event *binance.WsPartialDepthEvent) {
		timeStr := time.Now().Format("2006-01-02 15:04:05")

		// 获取 level 价格, 转换成字符串数组
		askbid := []string{
			event.Bids[bw.config.Level].Price,
			event.Asks[bw.config.Level].Price,
		}
		askbidBytes, _ := json.Marshal(askbid)
		askbidStr := string(askbidBytes)

		fmt.Println(timeStr, event.Symbol, "ask bid price", askbidStr)

		err := bw.redisClient.ZAdd(
			"z_askbid_binance_"+event.Symbol,
			&redis.Z{
				Score:  float64(time.Now().UnixNano() / 1e6),
				Member: askbidStr,
			}).Err()
		if err != nil {
			fmt.Println("redis zadd ", "z_askbid_binance_"+event.Symbol, err)
			return
		}
	}

	errHandler := func(err error) {
		fmt.Println(err)
	}

	symbolDepth := make(map[string]string)
	for _, symbol := range bw.config.Symbols {
		symbolDepth[symbol] = string(bw.config.Depth)

		if bw.config.Interval != "" {
			symbolDepth[symbol] = string(bw.config.Depth + "@" + bw.config.Interval)
		}

		go func(symbol string) {
			ticker := time.NewTicker(time.Second * 5)
			bw.tickers = append(bw.tickers, ticker)

			for range ticker.C {
				err := bw.redisClient.ZRemRangeByScore(
					"z_askbid_binance_"+symbol,
					"0",
					strconv.FormatInt(time.Now().UnixNano()/1e6-60*60*1e3, 10),
				).Err()
				if err != nil {
					fmt.Println("redis zremrangebyscore", "z_askbid_binance_"+symbol, err)
				}
			}
		}(symbol)
	}

	var err error
	bw.depthDoneC, bw.depthStopC, err = binance.WsCombinedPartialDepthServe(symbolDepth, wsPartialDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Notify 观察者模式调用的通知方法
func (bw *BinanceWatcher) Notify() {
	fmt.Println("Binance Config Chanaged", bw.config)

	// bw.depthDoneC <- struct{}{}
	bw.depthStopC <- struct{}{}

	for _, ticker := range bw.tickers {
		fmt.Println("Ticker stops")
		ticker.Stop()
	}
	bw.tickers = []*time.Ticker{}

	bw.WatchDepth()
}
