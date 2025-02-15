// Package pkg 提供了通用的工具组件
package pkg

import (
	"sync"
)

// LogEntry 表示一条日志记录
// 包含时间戳和日志消息两个基本要素
type LogEntry struct {
	// Timestamp 是日志生成时的Unix纳秒时间戳
	// 使用纳秒级时间戳可以保证日志的精确排序
	Timestamp int64

	// Message 存储实际的日志内容
	// 可以是任意字符串消息
	Message string

	// Level 日志级别
	Level LogLevel
}

// Buffer 实现了一个线程安全的日志缓冲区
// 主要功能：
// 1. 临时存储待发送的日志
// 2. 支持批量操作
// 3. 确保并发安全
type Buffer struct {
	// entries 存储所有待发送的日志条目
	// 使用切片实现，支持动态增长
	entries []LogEntry

	// mu 互斥锁，用于保护并发访问
	// 确保在多个goroutine同时操作时的数据一致性
	mu sync.Mutex

	// size 表示缓冲区的目标大小
	// 当日志数量达到此大小时，应该触发发送操作
	size int
}

// NewBuffer 创建并初始化一个新的缓冲区
// 参数：
//   - size: 缓冲区的目标大小，达到此大小时应触发发送
//
// 返回：
//   - *Buffer: 初始化好的缓冲区实例
func NewBuffer(size int) *Buffer {
	return &Buffer{
		// 预分配切片，容量设置为目标大小
		// 这样可以减少动态扩容的次数，提高性能
		entries: make([]LogEntry, 0, size),
		size:    size,
	}
}

// Add 向缓冲区添加一条日志
// 该方法是线程安全的，可以被多个goroutine同时调用
// 参数：
//   - entry: 要添加的日志条目
//
// 返回：
//   - bool: 如果缓冲区达到目标大小返回true，表示应该触发发送操作
func (b *Buffer) Add(entry LogEntry) bool {
	// 加锁保护并发访问
	b.mu.Lock()
	// 确保在方法返回时解锁
	defer b.mu.Unlock()

	// 添加日志条目到切片
	b.entries = append(b.entries, entry)
	// 检查是否达到目标大小
	return len(b.entries) >= b.size
}

// Flush 清空缓冲区并返回所有日志条目
// 该方法是线程安全的，通常在需要发送日志时调用
// 返回：
//   - []LogEntry: 所有待发送的日志条目
//
// 说明：
//
//	调用此方法后，缓冲区会被清空，返回的切片包含所有之前的日志条目
func (b *Buffer) Flush() []LogEntry {
	// 加锁保护并发访问
	b.mu.Lock()
	// 确保在方法返回时解锁
	defer b.mu.Unlock()

	// 保存当前的日志条目
	entries := b.entries
	// 创建新的空切片，预分配容量以优化性能
	b.entries = make([]LogEntry, 0, b.size)
	// 返回之前的日志条目
	return entries
}
