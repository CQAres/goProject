package SyncEventMgr

import (
	"fmt"
	"goProject/EventMgr/config"
	"sync"
)

type EventEntry struct {
	listenerName string
	listener     config.EventListenerHandler
	next         *EventEntry
}

type SyncEventMgr struct {
	listenerList [config.EVENT_MAX]*EventEntry
}

var (
	instance *SyncEventMgr
	once     sync.Once
)

func GetSyncEventMgr() *SyncEventMgr {
	once.Do(func() {
		instance = &SyncEventMgr{}
	})
	return instance
}

func (e *EventEntry) Log(eventType int, err error) {
	fmt.Printf("eventType=%d, listenerName=%s, err=%v\n", eventType, e.listenerName, err)
}

func (mgr *SyncEventMgr) RegListener(eventType int, listenerName string, listener config.EventListenerHandler) error {
	if eventType < config.EVENT_MIN || eventType >= config.EVENT_MAX {
		return fmt.Errorf("eventType is invalid, eventType=%d, err=%d", eventType, config.ERR_EVENT_INVALID)
	}
	if listenerName == "" {
		return fmt.Errorf("listenerName is nil, err=%d", config.ERR_LISTENER_NIL)
	}
	if listener == nil {
		return fmt.Errorf("listener is nil, err=%d", config.ERR_LISTENER_NIL)
	}

	for entry := mgr.listenerList[eventType]; entry != nil; entry = entry.next {
		if entry.listener == listener {
			return fmt.Errorf("listenerName is already registered, listenerName=%s, err=%d", listenerName, config.ERR_LISTENER_DUP)
		}
	}
	newEntry := &EventEntry{
		listenerName: listenerName,
		listener:     listener,
		next:         mgr.listenerList[eventType],
	}

	mgr.listenerList[eventType] = newEntry

	return nil
}

func (mgr *SyncEventMgr) RmvListener(eventType int, listener config.EventListenerHandler) error {
	if eventType < config.EVENT_MIN || eventType >= config.EVENT_MAX {
		return fmt.Errorf("eventType is invalid, eventType=%d, err=%d", eventType, config.ERR_EVENT_INVALID)
	}
	if listener == nil {
		return fmt.Errorf("listener is nil, err=%d", config.ERR_LISTENER_NIL)
	}

	entry := mgr.listenerList[eventType]
	var prev *EventEntry
	for entry != nil {
		if entry.listener == listener {
			if prev == nil {
				mgr.listenerList[eventType] = entry.next
			} else {
				prev.next = entry.next
			}
			return nil
		}
		prev = entry
		entry = entry.next
	}
	return fmt.Errorf("listener is not found, err=%d", config.ERR_LISTENER_NOTFD)
}

func (mgr *SyncEventMgr) DispatchEvent(eventType int, param config.EventData) error {
	if eventType < config.EVENT_MIN || eventType >= config.EVENT_MAX {
		return fmt.Errorf("eventType is invalid, eventType=%d, err=%d", eventType, config.ERR_EVENT_INVALID)
	}
	if param == nil {
		return fmt.Errorf("param is nil, err=%d", config.ERR_PARAM_NIL)
	}

	var errs []error
	for entry := mgr.listenerList[eventType]; entry != nil; entry = entry.next {
		if err := entry.listener.OnEvent(eventType, param); err != nil {
			entry.Log(eventType, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("dispatch event %d failed: %v", eventType, errs)
	}

	return nil
}
