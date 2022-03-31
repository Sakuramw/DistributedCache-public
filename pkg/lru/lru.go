package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         //设定空间
	nBytes    int64                         //已用空间
	ll        *list.List                    //存储链表(双向)
	cache     map[string]*list.Element      //利用map，对链表进行快速索引
	onEvicted func(key string, value Value) //回调函数，使用的人自己实现，在移除旧数据的时候会触发
}

// entry 链表存储的元素
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	if ele := c.ll.Back(); ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.onEvicted != nil {
			c.onEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(k string, v Value) {
	if ele, ok := c.cache[k]; ok {
		//已存在，更新map和list的值
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(v.Len()) - int64(kv.value.Len())
		kv.value = v
	} else {
		//不存在，添加
		ele := c.ll.PushFront(&entry{
			key:   k,
			value: v,
		})
		c.cache[k] = ele
		c.nBytes += int64(v.Len()) + int64(len(k))
	}
	//如果超过空间就不断移除旧数据
	for c.nBytes > c.maxBytes && c.maxBytes != 0 {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
