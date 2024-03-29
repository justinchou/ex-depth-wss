package service

import (
	"ex-depth-wss/utils"
	"fmt"
	"strings"
	"time"
)

// BinanceConf 币安配置
type BinanceConf struct {
	Filename string

	Symbol   string   `json:"symbol"`
	Depth    string   `json:"depth"`
	Level    int      `json:"level"`
	Interval string   `json:"interval"`
	Symbols  []string `json:"symbols"`

	histSymbol   string
	histDepth    string
	histLevel    int
	histInterval string

	isChanged bool
	observers []Observer
}

// Init 初始化配置
func (conf *BinanceConf) Init(filename string) {
	conf.Filename = filename

	conf.ReadWatchConf()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for range ticker.C {
			conf.ReadWatchConf()
		}
	}()
}

// FormatSymbol 格式化 Symbol
func (conf *BinanceConf) FormatSymbol(symbol string) string {
	return strings.ToUpper(strings.Replace(symbol, "_", "", 1))
}

// FormatSymbols 格式化 Symbols
func (conf *BinanceConf) FormatSymbols(symbols []string) []string {
	var n = []string{}
	for _, symbol := range symbols {
		n = append(n, conf.FormatSymbol(symbol))
	}
	return n
}

// ReadWatchConf 读取监听盘口价格配置
func (conf *BinanceConf) ReadWatchConf() (err error) {
	iniParser := utils.IniParser{}
	if err := iniParser.Load(conf.Filename); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", conf.Filename, err.Error())
		return err
	}

	block := "binance"

	conf.Symbol = iniParser.GetString(block, "symbol")
	conf.Depth = iniParser.GetString(block, "depth")
	conf.Level = int(iniParser.GetInt32(block, "level"))
	conf.Interval = iniParser.GetString(block, "interval")
	conf.Symbols = conf.FormatSymbols(Uniq(strings.Split(conf.Symbol, ",")))

	if conf.Symbol != conf.histSymbol ||
		conf.Depth != conf.histDepth ||
		conf.Level != conf.histLevel ||
		conf.Interval != conf.histInterval {
		conf.isChanged = true
	}

	conf.histSymbol = conf.Symbol
	conf.histDepth = conf.Depth
	conf.histLevel = conf.Level
	conf.histInterval = conf.Interval

	if conf.isChanged == true {
		fmt.Println(block, "conf hot reload", conf)

		for _, observer := range conf.observers {
			observer.Notify()
		}
		conf.isChanged = false
	}

	return nil
}

// AddObserver 注册观察者
func (conf *BinanceConf) AddObserver(observer Observer) {
	conf.observers = append(conf.observers, observer)
}
