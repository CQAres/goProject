package RoleLogout

import (
	"goProject/EventMgr/config"
)

type LogoutEvent struct {
	RoleID   uint64
	ServerID uint32
	IP       string
}

func (e *LogoutEvent) GetEventType() int {
	return config.EVENT_ROLE_LOGOUT
}
