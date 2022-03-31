// Package consistenthash 一致性hash算法
package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 哈希函数类型
type Hash func(data []byte) uint32

type HashCtl struct {
	hash             Hash           //hash算法
	replicas         int            //虚拟节点数
	nodeRing         []int          //节点哈希值环,是个有序切片，按照升序排列着虚拟节点的哈希值
	virtualToRealMap map[int]string //虚拟节点与真实节点的映射表，键是虚拟节点的哈希值，值是真实节点的名称
}

func New(replicas int, hs Hash) *HashCtl {
	m := &HashCtl{
		hash:             hs,
		replicas:         replicas,
		virtualToRealMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加节点
func (m *HashCtl) Add(nodes ...string) {
	for _, nodeName := range nodes {
		for i := 0; i < m.replicas; i++ {
			//每个节点创建replicas个虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + nodeName)))
			m.nodeRing = append(m.nodeRing, hash)
			m.virtualToRealMap[hash] = nodeName
		}
	}
	sort.Ints(m.nodeRing)
}

// Get 根据key的hash来搜索到应该存放的节点
func (m *HashCtl) Get(key string) string {
	if len(m.nodeRing) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.nodeRing), func(i int) bool {
		return m.nodeRing[i] >= hash
	})
	//节点的hash都小于key的hash，顺时针环回到首节点
	return m.virtualToRealMap[m.nodeRing[idx%len(m.nodeRing)]]
}
