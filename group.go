package main

import (
	"DistributedCache/peers"
	"DistributedCache/pkg/cachepb"
	"errors"
	"golang.org/x/sync/singleflight"
	"sync"
)

var (
	ErrInvalidGetter = errors.New("getter is nil")
	ErrEmptyKey      = errors.New("key is required")
	ErrPickerExists  = errors.New("peer picker is already exists")
	ErrPickerEmpty   = errors.New("can`t get picker")
)

// Getter 缓存中没有的时候应该怎么办，Get回调函数
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// Get 接口型函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 暴露外层的存储结构封装
type Group struct {
	name              string
	getter            Getter              //group实现的方法，当缓存中没有时应该怎么做
	mainCache         cache               //封装好的可并发核心缓存
	peerPicker        peers.PeerPicker    //给group加入节点选择接口
	singleFlightGroup *singleflight.Group //限制重复请求
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) //groups的map索引
)

func NewGroup(name string, cacheSpace int64, getter Getter) (*Group, error) {
	if getter == nil {
		return nil, ErrInvalidGetter
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:              name,
		getter:            getter,
		mainCache:         cache{setCacheMaxBytes: cacheSpace},
		singleFlightGroup: new(singleflight.Group),
	}
	groups[name] = group
	return group, nil
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从该group中获取一个key 的value
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, ErrEmptyKey
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	//不在缓存
	return g.load(key)
}

//load 从本地或者其他节点获取value
func (g *Group) load(key string) (ByteView, error) {
	value, err, _ := g.singleFlightGroup.Do(key, func() (interface{}, error) {
		//time.Sleep(5 * time.Second)
		if g.peerPicker != nil {
			if peer, ok := g.peerPicker.PickPeer(key); ok {
				request := &cachepb.Request{
					Group: g.name,
					Key:   key,
				}
				bytes, err := peer.Get(request)
				if err != nil {
					return ByteView{}, err
				}
				return ByteView{bytes}, nil
			}
			//return ByteView{}, ErrPickerEmpty
		}
		return g.loadFromLocal(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return value.(ByteView), nil
}

//loadFromLocal 从本地获取value
func (g *Group) loadFromLocal(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	//pop进缓存
	g.mainCache.add(key, value)
	return value, nil
}

// RegisterPeers 注入实现了节点选择器的对象
func (g *Group) RegisterPeers(picker peers.PeerPicker) error {
	//已经存在节点选择器
	if g.peerPicker != nil {
		return ErrPickerExists
	}
	g.peerPicker = picker
	return nil
}
