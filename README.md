# 废弃仓库

# loki-client-go
loki go客户端

# 本项目功能
收集日志通过http协议推送到loki
作为基础库给其他项目使用
写一个main函数, 作为示例
调用客户端,写入日志时,不立即推送, 而是先缓存到内存中, 达到一定数量后, 再推送.

loki api文档地址
```
https://grafana.org.cn/docs/loki/latest/reference/loki-http-api/#ingest-logs
```

日志时间戳单位: 纳秒时间戳
推送方式: 通过http协议推送到loki
推送频率: 最短间隔1秒, 最长间隔10秒. 如果没有100条数据, 则等待10秒. 如果10秒内没有100条数据, 则立即推送. 已达到100条数据, 则立即推送,最低频率 1秒一次.
