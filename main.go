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
	binanceConf := &service.BinanceConf{}
	binanceConf.Init(filename)

	binanceWatcher := &service.BinanceWatcher{}
	binanceWatcher.Init(redisClient, binanceConf)

	binanceConf.AddObserver(binanceWatcher)

	binanceWatcher.WatchDepth()
}

func startWatchOKEx(filename string, redisClient *redis.Client) {
	okexConf := &service.OKExConf{}
	okexConf.Init(filename)

	okexWatcher := &service.OKExWatcher{}
	okexWatcher.Init(redisClient, okexConf)

	okexConf.AddObserver(okexWatcher)

	okexConf.AddObserver(okexWatcher)

	okexWatcher.WatchDepth()
}

func main() {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(timeStr, "start service")

	filename := "etc/conf.ini"

	redisClient, err := service.ConnectRedis()
	if err != nil {
		fmt.Println("redis connect failed", err)
		return
	}

	go func() {
		startWatchBinance(filename, redisClient)
	}()

	go func() {
		startWatchOKEx(filename, redisClient)
	}()

	//合建chan
	c := make(chan os.Signal)
	//监听所有信号
	signal.Notify(c)
	//阻塞直到有信号传入
	fmt.Println("启动")
	s := <-c
	fmt.Println("退出信号", s)
}
