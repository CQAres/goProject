package RoleLogin

import (
	"fmt"
	"goProject/EventMgr/config"
)

type LoginListener struct{}

func (l *LoginListener) OnEvent(iEventType int, param config.EventData) error {
	// param 是 EventData 接口，需要断言成具体类型才能拿到字段
	loginEvt, ok := param.(*LoginEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", param)
	}

	fmt.Printf("LoginListener: RoleID=%d, IP=%s, ServerID=%d\n",
		loginEvt.RoleID, loginEvt.IP, loginEvt.ServerID)
	return nil
}
