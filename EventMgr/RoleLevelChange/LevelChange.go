package RoleLevelChange

import (
	"fmt"
	"goProject/EventMgr/config"
)

// LevelChangeLog 升级日志
type LevelChangeLog struct{}

// RankListener 排行榜
type RankListener struct{}

// FuncOpenListener 系统功能开启
type FuncOpenListener struct{}

func (l *LevelChangeLog) OnEvent(iEventType int, param config.EventData) error {
	// param 是 EventData 接口，需要断言成具体类型才能拿到字段
	LevelChangeEvt, ok := param.(*LevelChangeEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", param)
	}

	fmt.Printf("LevelChangeLog: RoleID=%d, OleLevel=%d, NewLevel=%d\n",
		LevelChangeEvt.RoleID, LevelChangeEvt.OleLevel, LevelChangeEvt.NewLevel)

	return nil
}

func (l *RankListener) OnEvent(iEventType int, param config.EventData) error {
	// param 是 EventData 接口，需要断言成具体类型才能拿到字段
	LevelChangeEvt, ok := param.(*LevelChangeEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", param)
	}

	fmt.Printf("RankListener: RoleID=%d, OleLevel=%d, NewLevel=%d\n",
		LevelChangeEvt.RoleID, LevelChangeEvt.OleLevel, LevelChangeEvt.NewLevel)

	return nil
}

func (l *FuncOpenListener) OnEvent(iEventType int, param config.EventData) error {
	// param 是 EventData 接口，需要断言成具体类型才能拿到字段
	LevelChangeEvt, ok := param.(*LevelChangeEvent)
	if !ok {
		return fmt.Errorf("unexpected event type: %T", param)
	}

	fmt.Printf("FuncOpenListener: RoleID=%d, OleLevel=%d, NewLevel=%d\n",
		LevelChangeEvt.RoleID, LevelChangeEvt.OleLevel, LevelChangeEvt.NewLevel)

	return nil
}
