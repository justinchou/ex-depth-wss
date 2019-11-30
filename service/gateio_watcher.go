package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/justinchou/gateio-go-sdk-api"
)

// GateIOWatcher 币安 websocket 数据监听服务
type GateIOWatcher struct {
	redisClient *redis.Client
	config      *GateIOConf
	gateWss     *gateio.GateWSAgent
	tickers     []*time.Ticker
}

// Init 初始化依赖注入
func (gw *GateIOWatcher) Init(client *redis.Client, config *GateIOConf) {
	gw.redisClient = client
	gw.config = config
}

// GetGateIOWss 创建 Wss 链接
func (gw *GateIOWatcher) GetGateIOWss() *gateio.GateWSAgent {
	if gw.gateWss != nil {
		return gw.gateWss
	}

	config := &gateio.Config{}

	config.PublicEndpoint = "http://data.gateio.co/"
	config.PrivateEndpoint = "https://api.gateio.co/"
	config.WSEndpoint = "wss://ws.gate.io/v3"
	config.ApiKey = ""
	config.SecretKey = ""
	config.Passphrase = ""
	config.TimeoutSecond = 45
	config.IsPrint = true
	config.I18n = gateio.ENGLISH

	wss := &gateio.GateWSAgent{}
	err := wss.Start(config)
	if err != nil {
		fmt.Println("gateio wss connect failed", err)
	}

	return wss
}

// WatchDepth 监听盘口价格
func (gw *GateIOWatcher) WatchDepth() {
	client := gw.GetGateIOWss()

	channel := "depth.subscribe"

	symbolDepth := []interface{}{}
	for _, symbol := range gw.config.Symbols {
		symbolDepth = append(symbolDepth, []interface{}{symbol, gw.config.Depth, "0.00000001"})

		go func(symbol string) {
			ticker := time.NewTicker(time.Second * 5)
			gw.tickers = append(gw.tickers, ticker)

			for range ticker.C {
				err := gw.redisClient.ZRemRangeByScore(
					"z_askbid_gateio_"+symbol,
					"0",
					strconv.FormatInt(time.Now().UnixNano()/1e6-60*60*1e3, 10),
				).Err()
				if err != nil {
					fmt.Println("redis zremrangebyscore", "z_askbid_gateio_"+symbol, err)
				}
			}
		}(symbol)
	}

	fmt.Println(channel, symbolDepth)
	err := client.Subscribe(channel, symbolDepth, gateio.DefaultPrintData)
	if err != nil {
		fmt.Println("subscribe failed", err)
	}
}

// Notify 观察者模式调用的通知方法
func (gw *GateIOWatcher) Notify() {
	fmt.Println("GateIO Config Chanaged", gw.config)

	for _, ticker := range gw.tickers {
		fmt.Println("Ticker stops")
		ticker.Stop()
	}
	gw.tickers = []*time.Ticker{}

	gw.WatchDepth()
}
