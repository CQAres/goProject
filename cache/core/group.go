package core

import (
	"fmt"
	"goProject/cache/util"
	"log"
	"sync"
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     util.PeerPicker
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func RegisterGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("cache: getter is nil")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	defer mutex.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (util.ByteView, error) {
	if key == "" {
		return util.ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value util.ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (util.ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return util.ByteView{}, err
	}
	value := util.ByteView{B: util.CloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value util.ByteView) {
	g.mainCache.add(key, 5, value)
}
