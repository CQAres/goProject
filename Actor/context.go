package actor

import "time"

// Context 提供 Actor 运行时的上下文信息与能力
type Context interface {
	// Self 返回当前 Actor 的引用
	Self() ActorRef
	// Sender 返回消息发送者的引用（可能为 nil）
	Sender() ActorRef
	// Message 返回当前处理的消息
	Message() any
	// Tell 向目标 Actor 发送异步消息
	Tell(target ActorRef, message any)
	// Ask 向目标 Actor 发送同步消息并等待响应
	Ask(target ActorRef, message any, timeoutMs int64) (any, error)
	// Spawn 创建并启动一个子 Actor
	Spawn(name string, actor Actor) ActorRef
	// Stop 停止指定的子 Actor
	Stop(child ActorRef)
	// SetReceiveTimeout 设置接收超时，超时后发送 TimeoutTick 消息
	SetReceiveTimeout(d time.Duration)
	// CancelReceiveTimeout 取消接收超时
	CancelReceiveTimeout()
}

// TimeoutTick 超时消息类型
type TimeoutTick struct{}
