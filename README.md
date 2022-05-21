# WechatTalk

企业微信自定义机器人 Go API.

### 限流限制
每个机器人发送的消息不能超过20条/分钟。

# 支持的消息类型：

* text 类型
* markdown 类型
* image 类型
* news 类型
* file 类型

# Installation

    go get github.com/xiaoxuan6/wechat-talk

# Test

    go test -v ./...