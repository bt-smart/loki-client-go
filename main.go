package main

import (
	"fmt"
	"github.com/bt-smart/loki-client-go/loki"
	"github.com/bt-smart/loki-client-go/pkg"
)

func main() {
	// 创建客户端配置
	config := loki.ClientConfig{
		URL: "http://192.168.98.207:3100",
		Labels: map[string]string{
			"service_name": "loki-client-go-dev",
		},
		BatchSize:   100,           // 每100条日志发送一次
		MinWaitTime: 1,             // 最少等待1秒
		MaxWaitTime: 10,            // 最多等待10秒
		MinLevel:    pkg.LevelInfo, // 设置最低日志级别为Info
	}

	// 创建并启动客户端
	client := loki.NewClient(config)
	client.Start()
	defer client.Stop()

	// 示例：发送不同级别的测试日志
	for i := 0; i < 1000; i++ {
		// Debug级别的日志会被忽略
		client.Debug(fmt.Sprintf(" message %d", i))

		// Info及以上级别的日志会被发送
		client.Info(fmt.Sprintf(" message %d", i))
		client.Warn(fmt.Sprintf(" message %d", i))
		client.Error(fmt.Sprintf(" message %d", i))

		// 模拟每100ms产生一组日志
		//time.Sleep(time.Millisecond * 100)
	}
}
