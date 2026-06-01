package actor

// Envelope 消息信封，包装消息及其元数据
type Envelope struct {
	Message any
	Sender  ActorRef
	// ResponseChan 用于 Ask 模式的响应通道
	ResponseChan chan any
}

// Mailbox Actor 邮箱，存储待处理的消息
type Mailbox struct {
	messages chan Envelope
}

// NewMailbox 创建指定大小的邮箱
func NewMailbox(size int) *Mailbox {
	return &Mailbox{
		messages: make(chan Envelope, size),
	}
}

// Push 将消息推入邮箱（非阻塞）
func (m *Mailbox) Push(envelope Envelope) bool {
	select {
	case m.messages <- envelope:
		return true
	default:
		// TODO: 邮箱满时的处理策略（丢弃、扩容等）
		return false
	}
}

// Pop 从邮箱取出一条消息（阻塞）
func (m *Mailbox) Pop() Envelope {
	return <-m.messages
}

// TryPop 尝试非阻塞取出消息
func (m *Mailbox) TryPop() (Envelope, bool) {
	select {
	case env := <-m.messages:
		return env, true
	default:
		return Envelope{}, false
	}
}

// Len 返回邮箱中消息数量
func (m *Mailbox) Len() int {
	return len(m.messages)
}
