package utils

import (
	"fmt"
	"os"
	"time"
)

type Observer interface {
	Notify(*Config)
}

type Config struct {
	Parser         IniParser
	Filename       string
	lastModifyTime int64
	observers      []*Observer
}

func (c *Config) NewConfig(filename string) (config *Config, err error) {
	config = &Config{
		Parser:   IniParser{},
		Filename: filename,
	}

	go config.reload()

	return config, nil
}

func (c *Config) reload() {
	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		func() {
			f, err := os.Open(c.Filename)
			if err != nil {
				fmt.Printf("open file error:%s\n", err)
				return
			}
			defer f.Close()

			fileInfo, err := f.Stat()
			if err != nil {
				fmt.Printf("stat file error:%s\n", err)
				return
			}

			// 或取当前文件修改时间
			curModifyTime := fileInfo.ModTime().Unix()
			if curModifyTime > c.lastModifyTime {
				c.lastModifyTime = curModifyTime

				// 配置更新通知所有观察者
				for _, n := range c.observers {
					(*n).Notify(c)
				}
			}
		}()
	}
}

func (c *Config) AddObserver(observer *Observer) {
	c.observers = append(c.observers, observer)
}
