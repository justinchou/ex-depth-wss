package service

import (
	"ex-depth-wss/utils"
	"fmt"
	"strings"
	"time"
)

// OKExConf 币安配置
type OKExConf struct {
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
	histSymbols  []string

	isChanged bool
	observers []Observer
}

// Init 初始化配置
func (oc *OKExConf) Init(filename string) {
	oc.Filename = filename

	oc.ReadWatchConf()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for range ticker.C {
			oc.ReadWatchConf()
		}
	}()
}

// FormatSymbol 格式化 Symbol
func (oc *OKExConf) FormatSymbol(symbol string) string {
	return strings.ToUpper(strings.Replace(symbol, "_", "-", 1))
}

// FormatSymbols 格式化 Symbols
func (oc *OKExConf) FormatSymbols(symbols []string) []string {
	var n = []string{}
	for _, symbol := range symbols {
		n = append(n, oc.FormatSymbol(symbol))
	}
	return n
}

// ReadWatchConf 读取监听盘口价格配置
func (oc *OKExConf) ReadWatchConf() (err error) {
	iniParser := utils.IniParser{}
	if err := iniParser.Load(oc.Filename); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", oc.Filename, err.Error())
		return err
	}

	oc.Symbol = iniParser.GetString("okex", "symbol")
	oc.Depth = iniParser.GetString("okex", "depth")
	oc.Level = int(iniParser.GetInt32("okex", "level"))
	oc.Interval = iniParser.GetString("okex", "interval")
	oc.Symbols = oc.FormatSymbols(Uniq(strings.Split(oc.Symbol, ",")))

	if oc.Symbol != oc.histSymbol ||
		oc.Depth != oc.histDepth ||
		oc.Level != oc.histLevel ||
		oc.Interval != oc.histInterval {
		oc.isChanged = true
	}

	oc.histSymbol = oc.Symbol
	oc.histDepth = oc.Depth
	oc.histLevel = oc.Level
	oc.histInterval = oc.Interval

	if oc.isChanged == true {
		fmt.Println("okex conf hot reload", oc)

		for _, observer := range oc.observers {
			observer.Notify()
		}
		oc.isChanged = false
	}

	return nil
}

// AddObserver 注册观察者
func (oc *OKExConf) AddObserver(observer Observer) {
	oc.observers = append(oc.observers, observer)
}
