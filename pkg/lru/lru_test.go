package lru

import (
	"strconv"
	"testing"
)

//var lru *Cache

type String string

func (s String) Len() int {
	return len(s)
}

func TestCache_AddAndGet(t *testing.T) {
	lru := NewCache(int64(2000), nil)
	lru.Add("k1", String("v1"))
	if v, ok := lru.Get("k1"); !ok || string(v.(String)) != "v1" {
		t.Fatal("get failed")
	}
	if _, ok := lru.Get("k2"); ok {
		t.Fatal("get panic")
	}
	//fmt.Println("end of add ")
}

func TestAutoRemove(t *testing.T) {
	lru := NewCache(int64(len("k1")+len("111"))*2, nil)
	for i := 0; i < 10; i++ {
		lru.Add("k"+strconv.Itoa(i), String("111"))
	}
	t.Log("lru length:", lru.Len())
	//fmt.Println("lru length:", lru.Len())
}

func TestOnEvicted(t *testing.T) {
	rmCnt := 0
	lru := NewCache(int64(2000), func(s string, value Value) {
		rmCnt++
	})
	lru.Add("k1", String("v1"))
	lru.Add("k2", String("v2"))
	lru.RemoveOldest()
	lru.RemoveOldest()
	if rmCnt != 2 {
		t.Fatal("call Evicted error")
	}
	//fmt.Println("end of evicted ")
}
