package main

import (
	"fmt"
	"goProject/EventMgr/RoleLevelChange"
	"goProject/EventMgr/RoleLogin"
	"goProject/EventMgr/RoleLogout"
	"goProject/EventMgr/SyncEventMgr"
	"goProject/EventMgr/config"
	"math/rand"
	"strconv"
	"sync"
)

// Init 初始化事件监听模块,其中包含目前共有三种行为事件：登录、升级、登出
// 其中升级事件包含三个监听模块，包括日志模块、排行榜模块、功能开启模块
func Init() {
	mgr := SyncEventMgr.GetSyncEventMgr()

	var login RoleLogin.LoginListener
	var logout RoleLogout.LogoutListener
	var log RoleLevelChange.LevelChangeLog
	var rank RoleLevelChange.RankListener
	var open RoleLevelChange.FuncOpenListener

	err := mgr.RegListener(config.EVENT_ROLE_LOGIN, "Login", &login)
	if err != nil {
		fmt.Printf("listener init failed, event=%d, listen=%s", config.EVENT_ROLE_LOGIN, "Login")
	}

	err = mgr.RegListener(config.EVENT_ROLE_LOGOUT, "Logout", &logout)
	if err != nil {
		fmt.Printf("listener init failed, event=%d, listen=%s", config.EVENT_ROLE_LOGOUT, "Logout")
	}

	err = mgr.RegListener(config.EVENT_ROLE_LEVEL_CHANGE, "rank", &rank)
	if err != nil {
		fmt.Printf("listener init failed, event=%d, listen=%s", config.EVENT_ROLE_LEVEL_CHANGE, "rank")
	}

	err = mgr.RegListener(config.EVENT_ROLE_LEVEL_CHANGE, "open", &open)
	if err != nil {
		fmt.Printf("listener init failed, event=%d, listen=%s", config.EVENT_ROLE_LEVEL_CHANGE, "open")
	}

	err = mgr.RegListener(config.EVENT_ROLE_LEVEL_CHANGE, "log", &log)
	if err != nil {
		fmt.Printf("listener init failed, event=%d, listen=%s", config.EVENT_ROLE_LEVEL_CHANGE, "log")
	}
}

// SyncTest 同步事件通知机制测试,同一事件的监听模块按顺序执行
func SyncTest() {
	Init()

	mgr := SyncEventMgr.GetSyncEventMgr()
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			role := uint64(i) + 1
			ip := strconv.Itoa(rand.Int())
			server := uint32(rand.Int())

			// 登录
			param1 := &RoleLogin.LoginEvent{
				RoleID:   role,
				IP:       ip,
				ServerID: server,
			}
			err := mgr.DispatchEvent(config.EVENT_ROLE_LOGIN, param1)
			if err != nil {
				return
			}

			// 升级
			param2 := &RoleLevelChange.LevelChangeEvent{
				RoleID:   uint64(i) + 1,
				OleLevel: 1,
				NewLevel: 2,
			}
			err = mgr.DispatchEvent(config.EVENT_ROLE_LEVEL_CHANGE, param2)
			if err != nil {
				return
			}

			// 登出
			param3 := &RoleLogout.LogoutEvent{
				RoleID:   role,
				IP:       ip,
				ServerID: server,
			}
			err = mgr.DispatchEvent(config.EVENT_ROLE_LOGOUT, param3)
			if err != nil {
				return
			}
		}()
	}
	wg.Wait()
}

func main() {
	SyncTest()
}
