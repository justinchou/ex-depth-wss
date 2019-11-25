package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/justinchou/okex-go-sdk-api"
)

// OKExWatcher 币安 websocket 数据监听服务
type OKExWatcher struct {
	redisClient *redis.Client
	config      *OKExConf
	okexWss     *okex.OKWSAgent
}

// Init 初始化依赖注入
func (ow *OKExWatcher) Init(client *redis.Client, config *OKExConf) {
	ow.redisClient = client
	ow.config = config
}

// NewOKExWss 创建 Wss 链接
func NewOKExWss() *okex.OKWSAgent {
	config := &okex.Config{}

	config.Endpoint = "https://www.okex.com/"
	config.WSEndpoint = "wss://real.okex.com:8443/"
	config.ApiKey = ""
	config.SecretKey = ""
	config.Passphrase = ""
	config.TimeoutSecond = 45
	config.IsPrint = false
	config.I18n = okex.ENGLISH

	client := &okex.OKWSAgent{}
	err := client.Start(config)
	if err != nil {
		fmt.Println("okex wss connect failed", err)
		return nil
	}

	return client
}

// WatchDepth 监听盘口价格
func (ow *OKExWatcher) WatchDepth() {
	timeStr := time.Now().Format("2006-01-02 15:04:05")

	client := NewOKExWss()
	channel := "spot/depth" + ow.config.Depth

	type ReceivedDataCallback func(interface{}) error
	receivedDataCallback := func(obj interface{}) error {
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
						Score:  float64(time.Now().Unix()),
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

	for _, symbol := range ow.config.Symbols {
		filter := symbol
		client.Subscribe(channel, filter, receivedDataCallback)
	}
}

// Notify 观察者模式调用的通知方法
func (bw *OKExWatcher) Notify() {
	fmt.Println("OKEx Config Chanaged", bw.config)

	bw.okexWss.Stop()

	// bw.WatchDepth()
}
