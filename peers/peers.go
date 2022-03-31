// Package peers 对节点选择和从节点中获取value的抽象
package peers

import "DistributedCache/pkg/cachepb"

// PeerPicker 交给服务器实现
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 交给客户端实现
type PeerGetter interface {
	Get(in *cachepb.Request) ([]byte, error)
	//Get(in *cachepb.Request,out *cachepb.Response) error
	//Get(group, key string) ([]byte, error)
	//Get(in *cachepb.Request) (*cachepb.Response, error)

}
