# DistributedCache-public

### 一个分布式缓存



#### 本地测试

编译

```shell
go build -o DC.exe ./
```

运行

```shell
DC.exe -p 1 
```



```shell
DC.exe -p 2
```



```shell
DC.exe -p 3 -b true
```

-p 代表开启cache 守护进程，与其他节点机器交互

测试中 1代表端口 8001，2代表8002，3代表8003

-b true/false 代表是否在此程序运行一个web api 交互程序 默认端口为9999 

模拟数据库的map

```go
var db = map[string]string{
   "Tom":  "111",
   "Jack": "222",
   "Sam":  "333",
}
```



开启两个个守护进程和一个守护与管理进程

访问:

```
http://127.0.0.1:9999/api?key=Tom
```

来测试
