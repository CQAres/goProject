package actor

// ==================== 系统消息 ====================

// PoisonPill 毒药消息，Actor 收到后自动停止
type PoisonPill struct{}

// Terminated Actor 终止通知消息
type Terminated struct {
	Who ActorRef
}

// Restarting Actor 正在重启通知
type Restarting struct{}

// Started Actor 已启动通知
type Started struct{}

// Stopped Actor 已停止通知
type Stopped struct{}

// ==================== 常用消息类型 ====================

// Ping Ping 消息
type Ping struct{}

// Pong Pong 响应
type Pong struct{}

// GenericMessage 通用消息包装器
type GenericMessage struct {
	Type    string
	Payload any
}
