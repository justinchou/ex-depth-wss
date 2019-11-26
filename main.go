package main

import (
	"ex-depth-wss/service"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v7"
)

func startWatchBinance(filename string, redisClient *redis.Client) {
	// 获取配置文件
	binanceConf := &service.BinanceConf{}
	binanceConf.Init(filename)

	// 初始化监听
	binanceWatcher := &service.BinanceWatcher{}
	binanceWatcher.Init(redisClient, binanceConf)

	// 注册观察者
	binanceConf.AddObserver(binanceWatcher)

	// 开启盘口价格抓取
	binanceWatcher.WatchDepth()
}

func startWatchOKEx(filename string, redisClient *redis.Client) {
	// 获取配置文件
	okexConf := &service.OKExConf{}
	okexConf.Init(filename)

	// 初始化监听
	okexWatcher := &service.OKExWatcher{}
	okexWatcher.Init(redisClient, okexConf)

	// 注册观察者
	okexConf.AddObserver(okexWatcher)

	// 开启盘口价格抓取
	okexWatcher.WatchDepth()
}

func main() {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(timeStr, "start service")

	filename := "etc/conf.ini"

	// 链接 Redis 数据库
	redisClient, err := service.ConnectRedis()
	if err != nil {
		fmt.Println("redis connect failed", err)
		return
	}

	go func() {
		// 抓取 Binance 盘口价格
		startWatchBinance(filename, redisClient)
	}()

	go func() {
		// 抓取 OKEx 盘口价格
		startWatchOKEx(filename, redisClient)
	}()

	c := make(chan os.Signal)
	// 监听所有信号
	signal.Notify(c)
	// 阻塞直到有信号传入
	fmt.Println("Service Started")
	s := <-c
	fmt.Println("Exit Signal", s)
}
