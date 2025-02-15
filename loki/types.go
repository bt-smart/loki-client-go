package loki

// Stream 表示一个日志流
// 包含流的标签信息和具体的日志内容
type Stream struct {
	// Stream 存储标签键值对，如 {"app": "myapp", "env": "prod"}
	Stream map[string]string `json:"stream"`
	// Values 存储日志记录，每条记录是一个长度为2的数组
	// [0]是时间戳字符串，[1]是日志消息
	Values [][2]string `json:"values"`
}

// PushRequest 表示向Loki发送的推送请求
// 一个请求可以包含多个日志流
type PushRequest struct {
	// Streams 包含所有要推送的日志流
	Streams []Stream `json:"streams"`
}

// ClientConfig 定义Loki客户端的配置参数
type ClientConfig struct {
	// URL 是Loki服务器的地址，例如 "http://localhost:3100"
	URL string // Loki服务器地址
	// Labels 定义默认的标签集，所有日志都会带上这些标签
	Labels map[string]string // 默认标签
	// BatchSize 定义批量发送的日志数量，达到这个数量就会触发发送
	BatchSize int // 批量发送的大小，默认100
	// MinWaitTime 定义两次发送之间的最小等待时间（秒）
	MinWaitTime int64 // 最小等待时间(秒)，默认1
	// MaxWaitTime 定义强制发送的最大等待时间（秒）
	MaxWaitTime int64 // 最大等待时间(秒)，默认10
}
