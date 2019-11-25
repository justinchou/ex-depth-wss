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

	isChanged bool
	observers []Observer
}

// Init 初始化配置
func (bc *OKExConf) Init(filename string) {
	bc.Filename = filename

	bc.ReadWatchConf()

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for _ = range ticker.C {
			bc.ReadWatchConf()
		}
	}()
}

// ReadWatchConf 读取监听盘口价格配置
func (bc *OKExConf) ReadWatchConf() (err error) {
	iniParser := utils.IniParser{}
	if err := iniParser.Load(bc.Filename); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", bc.Filename, err.Error())
		return err
	}

	bc.Symbol = iniParser.GetString("okex", "symbol")
	bc.Depth = iniParser.GetString("okex", "depth")
	bc.Level = int(iniParser.GetInt32("okex", "level"))
	bc.Interval = iniParser.GetString("okex", "interval")
	bc.Symbols = strings.Split(bc.Symbol, ",")

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
		fmt.Println("okex conf hot reload", bc)

		for _, observer := range bc.observers {
			observer.Notify()
		}
		bc.isChanged = false
	}

	return nil
}

// AddObserver 注册观察者
func (bc *OKExConf) AddObserver(observer Observer) {
	bc.observers = append(bc.observers, observer)
}
