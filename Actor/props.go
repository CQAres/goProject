package actor

// Props 用于配置和创建 Actor 实例
type Props struct {
	// Producer 创建 Actor 实例的工厂函数
	Producer func() Actor
	// MailboxSize 邮箱缓冲区大小，默认 1024
	MailboxSize int
	// SupervisionStrategy 监督策略
	SupervisionStrategy SupervisionStrategy
}

// NewProps 创建 Props
func NewProps(producer func() Actor) *Props {
	return &Props{
		Producer:            producer,
		MailboxSize:         1024,
		SupervisionStrategy: DefaultSupervisionStrategy,
	}
}

// WithMailboxSize 设置邮箱大小
func (p *Props) WithMailboxSize(size int) *Props {
	p.MailboxSize = size
	return p
}

// WithSupervisionStrategy 设置监督策略
func (p *Props) WithSupervisionStrategy(strategy SupervisionStrategy) *Props {
	p.SupervisionStrategy = strategy
	return p
}
