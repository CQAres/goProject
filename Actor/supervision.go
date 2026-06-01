package actor

// SupervisionDirective 监督指令
type SupervisionDirective int

const (
	// ResumeDirective 恢复 Actor，忽略当前消息
	ResumeDirective SupervisionDirective = iota
	// RestartDirective 重启 Actor
	RestartDirective
	// StopDirective 停止 Actor
	StopDirective
	// EscalateDirective 上报给父 Actor 处理
	EscalateDirective
)

// SupervisionStrategy 监督策略接口，决定 Actor 出错时的处理方式
type SupervisionStrategy interface {
	// Handle 决定如何处理子 Actor 的错误
	Handle(err error, child ActorRef) SupervisionDirective
}

// DefaultSupervisionStrategy 默认监督策略：始终重启
var DefaultSupervisionStrategy SupervisionStrategy = &OneForOneStrategy{MaxRetries: 3}

// OneForOneStrategy 一对一策略：只处理出错的 Actor
type OneForOneStrategy struct {
	MaxRetries    int                              // 最大重启次数
	RetriesTimeMs int64                            // 重启计数窗口（毫秒）
	Decider       func(error) SupervisionDirective // 错误决策函数
}

func (s *OneForOneStrategy) Handle(err error, child ActorRef) SupervisionDirective {
	if s.Decider != nil {
		return s.Decider(err)
	}
	// TODO: 实现 MaxRetries 计数逻辑，超过最大重启次数则停止
	return RestartDirective
}

// AllForOneStrategy 全对一策略：一个子 Actor 出错，所有子 Actor 都受影响
type AllForOneStrategy struct {
	MaxRetries    int
	RetriesTimeMs int64
	Decider       func(error) SupervisionDirective
}

func (s *AllForOneStrategy) Handle(err error, child ActorRef) SupervisionDirective {
	// TODO: 实现全对一策略，影响所有兄弟 Actor
	if s.Decider != nil {
		return s.Decider(err)
	}
	return RestartDirective
}
