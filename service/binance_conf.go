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
func (bc *BinanceConf) Init(filename string) {
	bc.Filename = filename

	bc.ReadWatchConf()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for _ = range ticker.C {
			bc.ReadWatchConf()
		}
	}()
}

// FormatSymbol 格式化 Symbol
func (bc *BinanceConf) FormatSymbol(symbol string) string {
	return strings.ToUpper(strings.Replace(symbol, "_", "", 1))
}

// FormatSymbols 格式化 Symbols
func (bc *BinanceConf) FormatSymbols(symbols []string) []string {
	var n = []string{}
	for _, symbol := range symbols {
		n = append(n, bc.FormatSymbol(symbol))
	}
	return n
}

// ReadWatchConf 读取监听盘口价格配置
func (bc *BinanceConf) ReadWatchConf() (err error) {
	iniParser := utils.IniParser{}
	if err := iniParser.Load(bc.Filename); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", bc.Filename, err.Error())
		return err
	}

	bc.Symbol = iniParser.GetString("binance", "symbol")
	bc.Depth = iniParser.GetString("binance", "depth")
	bc.Level = int(iniParser.GetInt32("binance", "level"))
	bc.Interval = iniParser.GetString("binance", "interval")
	bc.Symbols = bc.FormatSymbols(Uniq(strings.Split(bc.Symbol, ",")))

	if bc.Symbol != bc.histSymbol ||
		bc.Depth != bc.histDepth ||
		bc.Level != bc.histLevel ||
		bc.Interval != bc.histInterval {
		bc.isChanged = true
	}

	bc.histSymbol = bc.Symbol
	bc.histDepth = bc.Depth
	bc.histLevel = bc.Level
	bc.histInterval = bc.Interval

	if bc.isChanged == true {
		fmt.Println("binance conf hot reload", bc)

		for _, observer := range bc.observers {
			observer.Notify()
		}
		bc.isChanged = false
	}

	return nil
}

// AddObserver 注册观察者
func (bc *BinanceConf) AddObserver(observer Observer) {
	bc.observers = append(bc.observers, observer)
}
