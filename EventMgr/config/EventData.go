package config

// EventData 使用接口实现事件数据的抽象化,业务方只需要实现自己的事件数据结构,然后实现GetEventType方法即可
type EventData interface {
	GetEventType() int
}

// EventListenerHandler 使用接口实现监听模块的抽象化,业务方只需要实现自己的事件事件处理函数OnEvent即可
type EventListenerHandler interface {
	OnEvent(iEventType int, param EventData) error
}

const (
	EVENT_MIN               = 0
	EVENT_ROLE_LOGIN        = 1
	EVENT_ROLE_LEVEL_CHANGE = 2
	EVENT_ROLE_LOGOUT       = 3
	EVENT_MAX               = 10
)

const (
	ERR_EVENT_INVALID  = 1001 // 事件类型越界
	ERR_LISTENER_NIL   = 1002 // listener 为空
	ERR_LISTENER_DUP   = 1003 // listener 已注册
	ERR_LISTENER_NOTFD = 1004 // listener 未找到
	ERR_PARAM_NIL      = 1005 // param 为空
)
