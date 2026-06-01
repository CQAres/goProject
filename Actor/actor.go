package actor

// Actor 是所有 Actor 必须实现的接口
type Actor interface {
	// Receive 处理接收到的消息，返回响应（可选）
	Receive(ctx Context)
}

// ActorRef 是 Actor 的引用，用于向 Actor 发送消息
type ActorRef interface {
	// Tell 发送异步消息（fire-and-forget）
	Tell(message any)
	// Ask 发送同步消息并等待响应
	Ask(message any, timeoutMs int64) (any, error)
	// Path 返回 Actor 的路径
	Path() string
	// Stop 停止该 Actor
	Stop()
}
