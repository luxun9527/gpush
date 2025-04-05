# 高性能websocket推送组件


## ws推送优化

实现了这些优化。

1、写多读少的情况，使用epoll io多路复用避免创建读的协程。

2、分片存储连接，降低锁的粒度，提高推送的并发。

3、写的时候使用缓存，批量定时写入，减少系统调用、协程的调度。

![image.png](https://cdn.nlark.com/yuque/0/2025/png/12466223/1743862861672-38882091-085b-4d5b-9bdb-694955f072de.png?x-oss-process=image%2Fformat%2Cwebp)

### 1、epoll

1、实现参考，https://github.com/eranyanay/1m-go-websockets   只适用在写多读少的情况下，并且读之后的业务耗时的操作不要太频繁，可以使用epoll避免创建读的协程。

这里读的时候要考虑到半消息的情况，https://studygolang.com/topics/13377  当收到websocket数据包不全的情况下，go net官方库当不可读的时候是会阻塞的。如果这里如果我们也阻塞的的话，那就要等他读完我们才能处理其他的。

**解决方法**：如果如果没有读完的，就缓存起来，读的起始位置就回退到上次读出完整数据的位置，这里需要自己diy bufio.Reader 缓存没有读完的消息。

**除非你的单个实例的连接数非常大，可以做这个优化，否则感觉每啥必要。**



![image.png](https://cdn.nlark.com/yuque/0/2025/png/12466223/1743862901436-b10f6e9d-435c-45c5-8e9a-9e314e6f463d.png?x-oss-process=image%2Fformat%2Cwebp%2Fresize%2Cw_898%2Climit_0)





### 2、使用写缓存

压力测试，在本地没有带宽限制的情况，我们的瓶颈是cpu,我们通过pprof分析，是直接写入到连接那一步cpu消耗大，如果在小包特别多得情况下，直接发送会有很多系统调用,gmp模型，每一次系统调用go都会有一次调度 会消耗cpu。在实时性要求不高的情况下，我们可以批量定时写入。

得益于github.com/gobwas/ws 提供的非常灵活的用法，我们可以暂时将我们要写的数据转为websocket数据包缓存起来， 批量定时写入。此外提前encode websocket数据包，也可以避免每一个连接发送的时候都去encode 一遍websocket数据包,提高效率。

**实际测试5000连接，启用批量推送，在cpu负载满的情况下，性能有较大提升，后面是开启批量推送**

![image.png](https://cdn.nlark.com/yuque/0/2025/png/12466223/1743862923713-e1abd842-a62d-464a-9e51-fe707ed81b6c.png?x-oss-process=image%2Fformat%2Cwebp)

#### 2.1启用压缩

启用压缩客户端的cpu消耗会增加，对服务器没有影响，服务端的数据是提前处理的，不用每次推送都去执行压缩，由于压力测试是在自己的客户端服务端都是跑在自己的电脑上



![image.png](https://cdn.nlark.com/yuque/0/2025/png/12466223/1743862955829-2e736239-5a7f-4978-8898-1a535fd4c82c.png?x-oss-process=image%2Fformat%2Cwebp%2Fresize%2Cw_898%2Climit_0)



### 3、分片存储连接

主要是参考https://github.com/owenliang/go-push

主要的思路就是分片存储连接，减少锁的粒度和推送的并发。



## 编译运行

本程序依赖etcd需要提前搭建etcd服务在配置文件中进行配置。

编译`make build`  或者使用docker编译 `make dockerBuild`

运行 `make run` 或docker 运行 `make dockerRun`

## 实现功能

- 推送数据到公共频道


- 推送数据到所有连接


- 推送数据到私有频道

### 保持连接

需要在规定时间内发送ping字符串，由于浏览器不支持发送PING数据帧，统一发送ping字符串所有操作都为Text类型。

### 订阅操作

客户端请求数据的基本结构

请求地址 ws://192.168.2.99:9992/ws

```json
{
    "code":1,
    "topic":"test",
    "data":"any"
}
```

code:表示操作的类型

- 1订阅公共频道。
- 2取消订阅公共频道。
- 3登录操作，需要你自己实现登录的逻辑。
- 4登出操作,需要你自己实现登出的逻辑，会退出你订阅的所有私有频道。
- 5订阅私有频道，需要登录。
- 6取消订阅私有频道，需要登录。

topic:你订阅的频道的名称

data:任意类型，保留字段，如果是登录操作你可以用这个字段传你的登录凭证信息。

### 后台推送

后台推送提供http接口和grpc接口

**http 接口**：使用grpc-gateway 提供

POST /v1/pushData

json格式数据

```json
{
    "uid":"123456",
    "topic":"test",
    "data":"5pWw5o2u"
}
```

uid    :  用户id当推送给私有频道指定用户的id，否则为空。

topic : 频道的名称，当你想推送给所有的时候为空。

data : 具体的推送的内容，推送的数据要经过base64编码

**grpc接口** github.com/luxun9527/gpush/proto

protobuf 格式

```protobuf
message Data{
  string uid =1;
  string topic=2;
  bytes data=3;
}
```
