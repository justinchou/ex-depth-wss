package service

import (
	"ex-depth-wss/utils"
	"fmt"

	"github.com/go-redis/redis/v7"
)

// 读取 redis 配置文件
func readRedisConf() (conf *redis.Options, err error) {
	iniParser := utils.IniParser{}
	confFilename := "etc/conf.ini"

	if err := iniParser.Load(confFilename); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", confFilename, err.Error())
		return nil, err
	}

	fmt.Println(iniParser.GetString("redis", "host") + ":" + iniParser.GetString("redis", "port"))

	return &redis.Options{
		Addr:     iniParser.GetString("redis", "host") + ":" + iniParser.GetString("redis", "port"),
		Password: iniParser.GetString("redis", "auth"),   // no password set
		DB:       int(iniParser.GetInt32("redis", "db")), // use default DB
	}, nil
}

// 链接 redis
func ConnectRedis() (client *redis.Client, err error) {
	conf, err := readRedisConf()
	if err != nil {
		fmt.Println("read redis conf failed", err)
		return nil, err
	}

	client = redis.NewClient(conf)
	if err != nil {
		fmt.Println("redis connect failed", err)
		return nil, err
	}

	return client, nil
}
