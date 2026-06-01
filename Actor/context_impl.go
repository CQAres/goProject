package actor

import "time"

// actorContext Context 接口的实现
type actorContext struct {
	self    ActorRef
	sender  ActorRef
	message any
	system  *ActorSystem
}

func (ctx *actorContext) Self() ActorRef {
	return ctx.self
}

func (ctx *actorContext) Sender() ActorRef {
	return ctx.sender
}

func (ctx *actorContext) Message() any {
	return ctx.message
}

func (ctx *actorContext) Tell(target ActorRef, message any) {
	target.Tell(message)
}

func (ctx *actorContext) Ask(target ActorRef, message any, timeoutMs int64) (any, error) {
	return target.Ask(message, timeoutMs)
}

func (ctx *actorContext) Spawn(name string, actor Actor) ActorRef {
	return ctx.system.spawnChild(ctx.self, name, actor)
}

func (ctx *actorContext) Stop(child ActorRef) {
	child.Stop()
	// TODO: 从父 Actor 的 children map 中移除
}

func (ctx *actorContext) SetReceiveTimeout(d time.Duration) {
	// TODO: 实现接收超时定时器
	// 启动一个 goroutine，在 d 时间内没有收到消息则发送 TimeoutTick
}

func (ctx *actorContext) CancelReceiveTimeout() {
	// TODO: 取消接收超时定时器
}
