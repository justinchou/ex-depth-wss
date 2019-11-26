package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/justinchou/okex-go-sdk-api"
)

// OKExWatcher 币安 websocket 数据监听服务
type OKExWatcher struct {
	redisClient *redis.Client
	config      *OKExConf
	okexWss     *okex.OKWSAgent
	channels    map[string][]string
	tickers     []*time.Ticker
}

// Init 初始化依赖注入
func (ow *OKExWatcher) Init(client *redis.Client, config *OKExConf) {
	ow.redisClient = client
	ow.config = config
}

// GetOKExWss 创建 Wss 链接
func (ow *OKExWatcher) GetOKExWss() *okex.OKWSAgent {
	if ow.okexWss != nil {
		return ow.okexWss
	}

	config := &okex.Config{}

	config.Endpoint = "https://www.okex.com/"
	config.WSEndpoint = "wss://real.okex.com:8443/"
	config.ApiKey = ""
	config.SecretKey = ""
	config.Passphrase = ""
	config.TimeoutSecond = 45
	config.IsPrint = false
	config.I18n = okex.ENGLISH

	ow.okexWss = &okex.OKWSAgent{}
	err := ow.okexWss.Start(config)
	if err != nil {
		fmt.Println("okex wss connect failed", err)
	}

	ow.channels = make(map[string][]string)
	return ow.okexWss
}

// WatchDepth 监听盘口价格
func (ow *OKExWatcher) WatchDepth() {
	client := ow.GetOKExWss()
	channel := "spot/depth" + ow.config.Depth

	type ReceivedDataCallback func(interface{}) error
	receivedDataCallback := func(obj interface{}) error {
		timeStr := time.Now().Format("2006-01-02 15:04:05")

		switch obj.(type) {
		case string:
			fmt.Println("recv string", obj)
		case int:
			fmt.Println("recv int", obj)
		case *okex.WSDepthTableResponse:
			obj, ok := obj.(*okex.WSDepthTableResponse)
			if !ok {
				return nil
			}

			for _, event := range obj.Data {
				// fmt.Println("recv depth", event.Timestamp, event.InstrumentId, event.Asks, event.Bids)

				// 获取 level 价格, 转换成字符串数组
				askbid := []string{
					event.Bids[ow.config.Level][0].(string),
					event.Asks[ow.config.Level][0].(string),
				}
				askbidBytes, _ := json.Marshal(askbid)
				askbidStr := string(askbidBytes)

				fmt.Println(timeStr, event.InstrumentId, "ask bid price", askbidStr)

				err := ow.redisClient.ZAdd(
					"z_askbid_okex_"+event.InstrumentId,
					&redis.Z{
						Score:  float64(time.Now().UnixNano() / 1e6),
						Member: askbidStr,
					}).Err()
				if err != nil {
					fmt.Println("redis zadd ", "z_askbid_okex_"+event.InstrumentId, err)
					return nil
				}
			}
		default:
			msg, err := okex.Struct2JsonString(obj)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
			fmt.Println("recv", msg)
		}
		return nil
	}

	if ow.channels[channel] == nil {
		ow.channels[channel] = []string{}
	}

	// UnSubscribe 直接将整个 channel 都停止监听了
	for _, symbol := range ow.channels[channel] {
		contain := Contains(ow.config.Symbols, symbol)

		if !contain {
			err := client.UnSubscribe(channel, symbol)
			fmt.Println("unsubscribe", channel, err)

			ow.channels[channel] = []string{}

			for _, ticker := range ow.tickers {
				fmt.Println("Ticker stops")
				ticker.Stop()
			}
			ow.tickers = []*time.Ticker{}

			break
		}
	}

	for _, symbol := range ow.config.Symbols {
		contain := Contains(ow.channels[channel], symbol)

		if !contain {
			ow.channels[channel] = append(ow.channels[channel], symbol)

			err := client.Subscribe(channel, symbol, receivedDataCallback)
			fmt.Println("subscribe", channel, symbol, err)

			go func() {
				ticker := time.NewTicker(time.Second * 5)
				ow.tickers = append(ow.tickers, ticker)

				for range ticker.C {
					err := ow.redisClient.ZRemRangeByScore(
						"z_askbid_okex_"+symbol,
						"0",
						strconv.FormatInt(time.Now().UnixNano()/1e6-60*60*1e3, 10),
					).Err()
					if err != nil {
						fmt.Println("redis zremrangebyscore", "z_askbid_okex_"+symbol, err)
					}
				}
			}()
		}
	}
}

// Notify 观察者模式调用的通知方法
func (ow *OKExWatcher) Notify() {
	fmt.Println("OKEx Config Chanaged", ow.config)

	ow.WatchDepth()
}
