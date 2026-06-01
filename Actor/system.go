package actor

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// ActorSystem Actor 系统，管理所有 Actor 的生命周期
type ActorSystem struct {
	name     string
	actors   map[string]*localActorRef
	actorsMu sync.RWMutex
	running  atomic.Bool
	root     ActorRef
	sequence int64
}

// NewActorSystem 创建 Actor 系统
func NewActorSystem(name string) *ActorSystem {
	sys := &ActorSystem{
		name:   name,
		actors: make(map[string]*localActorRef),
	}

	// 创建根 Guardian Actor
	sys.root = newLocalActorRef("/"+name, &guardianActor{}, NewMailbox(1024), sys)
	sys.actors[name] = sys.root.(*localActorRef)
	sys.running.Store(true)
	sys.root.(*localActorRef).start()

	return sys
}

// Root 返回根 Actor 引用
func (sys *ActorSystem) Root() ActorRef {
	return sys.root
}

// Spawn 创建并启动一个顶层 Actor
func (sys *ActorSystem) Spawn(name string, props *Props) ActorRef {
	return sys.spawnChild(sys.root, name, props.Producer())
}

// spawnChild 在指定父 Actor 下创建子 Actor
func (sys *ActorSystem) spawnChild(parent ActorRef, name string, actor Actor) ActorRef {
	path := parent.Path() + "/" + name

	sys.actorsMu.Lock()
	defer sys.actorsMu.Unlock()

	if _, exists := sys.actors[path]; exists {
		// Actor 已存在，返回错误或现有引用
		return sys.actors[path]
	}

	// TODO: 根据 Props 配置创建邮箱和设置监督策略
	ref := newLocalActorRef(path, actor, NewMailbox(1024), sys)
	sys.actors[path] = ref

	// 注册为父 Actor 的子 Actor
	if parentRef, ok := parent.(*localActorRef); ok {
		parentRef.childrenMu.Lock()
		parentRef.children[name] = ref
		parentRef.childrenMu.Unlock()
	}

	ref.start()
	return ref
}

// Stop 停止指定路径的 Actor
func (sys *ActorSystem) Stop(path string) {
	sys.actorsMu.Lock()
	defer sys.actorsMu.Unlock()

	if ref, exists := sys.actors[path]; exists {
		ref.Stop()
		delete(sys.actors, path)
	}
}

// Lookup 根据路径查找 Actor
func (sys *ActorSystem) Lookup(path string) (ActorRef, bool) {
	sys.actorsMu.RLock()
	defer sys.actorsMu.RUnlock()

	ref, exists := sys.actors[path]
	if exists {
		return ref, true
	}
	return nil, false
}

// Shutdown 关闭整个 Actor 系统
func (sys *ActorSystem) Shutdown() {
	sys.running.Store(false)

	sys.actorsMu.Lock()
	defer sys.actorsMu.Unlock()

	for _, ref := range sys.actors {
		ref.Stop()
	}
	sys.actors = make(map[string]*localActorRef)
}

// PrintTree 打印 Actor 层级树（调试用）
func (sys *ActorSystem) PrintTree() string {
	sys.actorsMu.RLock()
	defer sys.actorsMu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ActorSystem[%s]\n", sys.name))
	// TODO: 递归打印 Actor 树
	for path := range sys.actors {
		sb.WriteString(fmt.Sprintf("  - %s\n", path))
	}
	return sb.String()
}

// guardianActor 根 Guardian Actor，作为所有顶层 Actor 的父级
type guardianActor struct{}

func (g *guardianActor) Receive(ctx Context) {
	// Guardian Actor 默认忽略所有消息
	// TODO: 可实现顶层监督逻辑
}
