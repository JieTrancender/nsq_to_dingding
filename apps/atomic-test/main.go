package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// 加载配置信息，这里使用随机数表示获取配置
func loadConfig() map[string]int {
	m := make(map[string]int)

	m["id"] = rand.Int()

	return m
}

var request chan int

func requests() chan int {
	return request
}

func main() {
	fmt.Println("Hello World!")
	rand.Seed(int64(time.Now().UnixNano()))

	var config atomic.Value
	config.Store(loadConfig())

	go func() {
		for {
			// 每十秒钟定时拉取最新的配置信息
			time.Sleep(2 * time.Second)
			config.Store(loadConfig())
			c := config.Load().(map[string]int)
			fmt.Println("load newest config", c["id"])
		}
	}()

	// 创建工作协程，每个工作协程都会根据所读到的配置信息来处理请求
	for i := 0; i < 10; i++ {
		go func(a int) {
			for r := range requests() {
				c := config.Load().(map[string]int)
				fmt.Println("exec", a, c["id"])
				_ = r
			}
		}(i)
	}

	for i := 1; i < 100; i++ {
		time.Sleep(2 * time.Second)
		fmt.Println("~~~~~~~~request <- ", i)
		request <- i
	}
}
