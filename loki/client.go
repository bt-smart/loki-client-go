// Package loki 实现了Loki日志系统的客户端
package loki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bt-smart/loki-client-go/pkg"
)

// Client 实现了Loki的客户端，提供日志推送功能
// 支持批量发送、缓存、自动重试等特性
type Client struct {
	// config 存储客户端的配置信息，包括服务器地址、标签等
	config ClientConfig
	// buffer 是内存中的日志缓冲区，用于批量发送日志
	buffer *pkg.Buffer
	// done 是用于优雅关闭的信号通道
	done chan bool
}

// NewClient 创建并初始化一个新的Loki客户端实例
// 参数：
//   - config: 客户端配置，包含服务器地址、批量大小等设置
//
// 返回：
//   - *Client: 初始化好的客户端实例
func NewClient(config ClientConfig) *Client {
	// 设置默认的批量发送大小
	if config.BatchSize == 0 {
		config.BatchSize = 100 // 默认每100条日志发送一次
	}
	// 设置默认的最小等待时间
	if config.MinWaitTime == 0 {
		config.MinWaitTime = 1 // 默认最少等待1秒
	}
	// 设置默认的最大等待时间
	if config.MaxWaitTime == 0 {
		config.MaxWaitTime = 10 // 默认最多等待10秒
	}
	// 如果未指定最低日志级别，默认为 Info
	if config.MinLevel == 0 {
		config.MinLevel = pkg.LevelInfo
	}

	return &Client{
		config: config,
		buffer: pkg.NewBuffer(config.BatchSize),
		done:   make(chan bool),
	}
}

// Debug 记录调试级别的日志
func (c *Client) Debug(message string) error {
	return c.pushLogWithLevel(message, pkg.LevelDebug)
}

// Info 记录信息级别的日志
func (c *Client) Info(message string) error {
	return c.pushLogWithLevel(message, pkg.LevelInfo)
}

// Warn 记录警告级别的日志
func (c *Client) Warn(message string) error {
	return c.pushLogWithLevel(message, pkg.LevelWarn)
}

// Error 记录错误级别的日志
func (c *Client) Error(message string) error {
	return c.pushLogWithLevel(message, pkg.LevelError)
}

// pushLogWithLevel 内部方法，处理带级别的日志推送
func (c *Client) pushLogWithLevel(message string, level pkg.LogLevel) error {
	// 检查日志级别，低于最小级别的日志直接忽略
	if level < c.config.MinLevel {
		return nil
	}

	// 创建日志条目，使用纳秒级时间戳
	entry := pkg.LogEntry{
		Timestamp: time.Now().UnixNano(),
		Message:   message,
		Level:     level,
	}

	// 添加到缓冲区，如果缓冲区已满则触发发送
	if c.buffer.Add(entry) {
		c.flush()
	}
	return nil
}

// Start 启动客户端的后台工作协程
// 该协程负责定期检查并发送缓冲区中的日志
func (c *Client) Start() {
	go c.worker()
}

// Stop 停止客户端的后台工作协程
// 应在程序退出前调用，以确保所有日志都被发送
func (c *Client) Stop() {
	c.done <- true
}

// worker 是后台工作协程的主循环
// 负责定期检查并发送日志，实现了以下功能：
// 1. 定期检查是否需要发送日志
// 2. 处理优雅关闭信号
// 3. 确保日志不会在缓冲区中停留太久
func (c *Client) worker() {
	// 创建定时器，用于周期性检查是否需要发送日志
	ticker := time.NewTicker(time.Second * time.Duration(c.config.MaxWaitTime))
	lastFlush := time.Now()

	for {
		select {
		case <-c.done:
			// 收到关闭信号，退出工作协程
			return
		case <-ticker.C:
			// 检查是否超过最大等待时间
			if time.Since(lastFlush) >= time.Second*time.Duration(c.config.MaxWaitTime) {
				c.flush()
				lastFlush = time.Now()
			}
		}
	}
}

// flush 将缓冲区中的日志发送到Loki服务器
// 主要步骤：
// 1. 从缓冲区获取所有待发送的日志
// 2. 将日志转换为Loki期望的格式
// 3. 发送到服务器
func (c *Client) flush() {
	// 获取并清空缓冲区
	entries := c.buffer.Flush()
	if len(entries) == 0 {
		return
	}

	// 按日志级别分组
	levelGroups := make(map[pkg.LogLevel][][2]string)
	for _, entry := range entries {
		levelGroups[entry.Level] = append(levelGroups[entry.Level], [2]string{
			strconv.FormatInt(entry.Timestamp, 10),
			entry.Message,
		})
	}

	// 为每个级别创建单独的流
	var streams []Stream
	for level, values := range levelGroups {
		// 复制标签并添加级别
		labels := make(map[string]string)
		for k, v := range c.config.Labels {
			labels[k] = v
		}
		// 添加日志级别标签
		labels["detected_level"] = pkg.LevelToString(level)

		streams = append(streams, Stream{
			Stream: labels,
			Values: values,
		})
	}

	// 创建推送请求
	req := PushRequest{
		Streams: streams,
	}

	// 发送请求到Loki服务器
	err := c.send(req)
	if err != nil {
		log.Println(err.Error())
	}
}

// send 负责将日志请求发送到Loki服务器
// 参数：
//   - req: 要发送的日志请求
//
// 返回：
//   - error: 发送过程中的错误，如果成功则为nil
func (c *Client) send(req PushRequest) error {
	// 将请求序列化为JSON
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %v", err)
	}

	// 发送HTTP POST请求
	resp, err := http.Post(c.config.URL+"/loki/api/v1/push", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
