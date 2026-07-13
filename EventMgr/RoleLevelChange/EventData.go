package RoleLevelChange

import (
	"goProject/EventMgr/config"
)

type LevelChangeEvent struct {
	RoleID   uint64
	OleLevel uint32
	NewLevel uint32
}

func (e *LevelChangeEvent) GetEventType() int {
	return config.EVENT_ROLE_LEVEL_CHANGE
}
