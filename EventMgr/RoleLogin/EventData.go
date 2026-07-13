package RoleLogin

import (
	"goProject/EventMgr/config"
)

type LoginEvent struct {
	RoleID   uint64
	ServerID uint32
	IP       string
}

func (e *LoginEvent) GetEventType() int {
	return config.EVENT_ROLE_LOGIN
}
