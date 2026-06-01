package actor

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// localActorRef 本地 Actor 引用的实现
type localActorRef struct {
	path    string
	actor   Actor
	mailbox *Mailbox
	system  *ActorSystem

	running  atomic.Bool
	stopOnce sync.Once
	stopCh   chan struct{}

	// 子 Actor 管理
	children   map[string]ActorRef
	childrenMu sync.RWMutex

	// 监督相关
	restartCount int
}

func newLocalActorRef(path string, actor Actor, mailbox *Mailbox, system *ActorSystem) *localActorRef {
	return &localActorRef{
		path:     path,
		actor:    actor,
		mailbox:  mailbox,
		system:   system,
		stopCh:   make(chan struct{}),
		children: make(map[string]ActorRef),
	}
}

// start 启动 Actor 的消息处理循环
func (ref *localActorRef) start() {
	ref.running.Store(true)
	go ref.receiveLoop()
}

func (ref *localActorRef) receiveLoop() {
	for ref.running.Load() {
		select {
		case env := <-ref.mailbox.messages:
			ref.processMessage(env)
		case <-ref.stopCh:
			return
		}
	}
}

func (ref *localActorRef) processMessage(env Envelope) {
	// 捕获 panic，交给监督策略处理
	defer func() {
		if r := recover(); r != nil {
			// TODO: 构造 error 并交给 SupervisionStrategy 处理
			// strategy := ref.getSupervisionStrategy()
			// directive := strategy.Handle(err, ref)
			// ref.handleDirective(directive)
			_ = r
		}
	}()

	ctx := &actorContext{
		self:    ref,
		sender:  env.Sender,
		message: env.Message,
		system:  ref.system,
	}

	ref.actor.Receive(ctx)

	// 如果是 Ask 模式，关闭响应通道
	if env.ResponseChan != nil {
		close(env.ResponseChan)
	}
}

func (ref *localActorRef) Tell(message any) {
	if !ref.running.Load() {
		return
	}
	ref.mailbox.Push(Envelope{
		Message: message,
		Sender:  nil,
	})
}

func (ref *localActorRef) Ask(message any, timeoutMs int64) (any, error) {
	if !ref.running.Load() {
		return nil, fmt.Errorf("actor %s is stopped", ref.path)
	}

	respCh := make(chan any, 1)
	ref.mailbox.Push(Envelope{
		Message:      message,
		Sender:       ref.system.Root(),
		ResponseChan: respCh,
	})

	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		return nil, fmt.Errorf("ask timeout for actor %s", ref.path)
	}
}

func (ref *localActorRef) Path() string {
	return ref.path
}

func (ref *localActorRef) Stop() {
	ref.stopOnce.Do(func() {
		ref.running.Store(false)
		close(ref.stopCh)

		// 停止所有子 Actor
		ref.childrenMu.Lock()
		for _, child := range ref.children {
			child.Stop()
		}
		ref.childrenMu.Unlock()
	})
}
