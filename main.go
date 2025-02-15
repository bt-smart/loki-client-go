package main

import (
	"fmt"
	"github.com/bt-smart/loki-client-go/loki"
	"time"
)

func main() {
	config := loki.ClientConfig{
		URL:         "http://localhost:3100",
		Labels:      map[string]string{"job": "test", "env": "dev"},
		BatchSize:   100,
		MinWaitTime: 1,
		MaxWaitTime: 10,
	}

	client := loki.NewClient(config)
	client.Start()
	defer client.Stop()

	// 示例：发送一些测试日志
	for i := 0; i < 1000; i++ {
		client.PushLog(fmt.Sprintf("test log message %d", i))
		time.Sleep(time.Millisecond * 100)
	}
}
