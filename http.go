package main

import (
	"DistributedCache/peers"
	"DistributedCache/pkg/cachepb"
	"DistributedCache/pkg/consistenthash"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"sync"
)

var (
	ErrServerBusy = errors.New("服务繁忙")
)

type HttpServer struct {
	self        string //服务器自身地址
	basePath    string //通信前缀，默认为"/distributed_cache"
	mu          sync.Mutex
	nodesCtler  *consistenthash.HashCtl //由一致性哈希算法管理的节点
	httpClients map[string]*httpClient  //客户端组
}

func NewHttpServer(self string) *HttpServer {
	return &HttpServer{
		self:     self,
		basePath: "/distributed_cache",
	}
}

func (p *HttpServer) SetHttpService() *gin.Engine {
	r := gin.Default()

	r.GET(p.basePath+"/:group/:key", func(c *gin.Context) {
		groupName := c.Param("group")
		keyName := c.Param("key")
		group := GetGroup(groupName)
		if group == nil {
			resp, err := proto.Marshal(&cachepb.Response{Value: []byte("Group Name Is Invalid")})
			if err != nil {
				return
			}
			c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", resp)
			return
		}
		view, err := group.Get(keyName)
		if err != nil {
			resp, err1 := proto.Marshal(&cachepb.Response{Value: []byte(err.Error())})
			if err1 != nil {
				return
			}
			c.Data(http.StatusInternalServerError, "text/plain; charset=utf-8", resp)
			return
		}
		resp, err := proto.Marshal(&cachepb.Response{Value: view.Bytes()})
		if err != nil {
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", resp)
	})
	return r
}

//http客户端
type httpClient struct {
	urlPF string //访问节点的url前缀,e.g.: http://baidu.com/distributed_cache"
}

// Get 从远程节点获取value
func (g *httpClient) Get(in *cachepb.Request) ([]byte, error) {
	url := g.urlPF + "/" + in.GetGroup() + "/" + in.GetKey()
	res, err := http.Get(url)
	fmt.Println("get key from:", url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, ErrServerBusy
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	resp := new(cachepb.Response)
	err = proto.Unmarshal(bytes, resp)
	if err != nil {
		return nil, err
	}
	return resp.GetValue(), nil
}

//func (g *httpClient) Get(group, key string) ([]byte, error) {
//	url := g.urlPF + "/" + group + "/" + key
//	res, err := http.Get(url)
//	fmt.Println("get key from:", url)
//	if err != nil {
//		return nil, err
//	}
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK {
//		return nil, ErrServerBusy
//	}
//	bytes, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return nil, err
//	}
//	return bytes, nil
//}

//编译期验证HttpClient实现 PeerGetter接口
var _ peers.PeerGetter = (*httpClient)(nil)

const (
	defaultReplicas = 50
	defaultBasePath = "distributed_cache"
)

// Set 把分布式节点的地址添加进去
func (p *HttpServer) Set(nodes ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nodesCtler = consistenthash.New(defaultReplicas, nil) //创建一个一致性哈希节点管理
	p.nodesCtler.Add(nodes...)
	p.httpClients = make(map[string]*httpClient, len(nodes))
	for _, nodeUrl := range nodes {
		p.httpClients[nodeUrl] = &httpClient{urlPF: "http://" + nodeUrl + p.basePath}
	}
}

// PickPeer 让服务器实现了PeerPicker接口,可以根据传入的key找到getter方法(即应创建的客户端)
//从而启动客户端调用get方法获取到value
func (p *HttpServer) PickPeer(key string) (peers.PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.nodesCtler.Get(key); peer != "" && peer != p.self {
		return p.httpClients[peer], true
	}
	return nil, false
}

var _ peers.PeerPicker = (*HttpServer)(nil)
