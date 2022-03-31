//value抽象的实现，实现了len()方法
package main

type ByteView struct {
	b []byte //支持多种内容
}

func (v ByteView) Len() int {
	return len(v.b)
}

// Bytes 返回一个底层数组不一样的切片防止修改
func (v ByteView) Bytes() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (v ByteView) String() string {
	return string(v.b)
}
