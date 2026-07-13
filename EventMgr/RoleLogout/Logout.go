package RoleLogout

import (
	"fmt"
	"goProject/EventMgr/config"
)

type LogoutListener struct{}

func (l *LogoutListener) OnEvent(iEventType int, param config.EventData) error {
	// param 是 EventData 接口，需要断言成具体类型才能拿到字段
	logoutEvt, ok := param.(*LogoutEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", param)
	}

	fmt.Printf("LogoutListener: RoleID=%d, IP=%s, ServerID=%d\n",
		logoutEvt.RoleID, logoutEvt.IP, logoutEvt.ServerID)
	return nil
}
