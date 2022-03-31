package main

import (
	"errors"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

var db = map[string]string{
	"Tom":  "111",
	"Jack": "222",
	"Sam":  "333",
}

var (
	wg sync.WaitGroup
	ch = make(chan string)
)

func createGroup() *Group {
	group, _ := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key from slow database", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, errors.New("key dose not exists")
		}))
	return group
}

func startCacheServer(addr string, address []string, cache *Group) {
	server := NewHttpServer(addr)
	server.Set(address...)
	cache.RegisterPeers(server)
	log.Println("server is running at", addr)
	r := server.SetHttpService()
	r.Run(addr)
}

func startCacheBoard(boardAddress string, gp *Group) {
	r := gin.Default()
	r.GET("/api", func(c *gin.Context) {
		//cfc := c.Copy()
		//go func() {
		//	key := cfc.Query("key")
		//	bv, err := gp.Get(key)
		//	if err != nil {
		//		cfc.String(http.StatusOK, key+"不存在")
		//		return
		//	}
		//	ch<- bv.String()
		//}()
		//c.String(http.StatusOK,<-ch)
		key := c.Query("key")
		bv, err := gp.Get(key)
		if err != nil {
			c.String(http.StatusOK, key+"不存在")
			return
		}
		c.String(http.StatusOK, bv.String())

	})
	r.Run(boardAddress)
}

func main() {
	var port int
	var runBoard bool
	flag.IntVar(&port, "p", 0, "server port")
	flag.BoolVar(&runBoard, "b", false, "Start a board")
	flag.Parse()
	boardAddr := "127.0.0.1:9999"
	addrs := []string{
		"",
		"127.0.0.1:8001",
		"127.0.0.1:8002",
		"127.0.0.1:8003",
	}

	caches := createGroup()
	if runBoard {
		go startCacheBoard(boardAddr, caches)
	}
	startCacheServer(addrs[port], addrs, caches)

}
