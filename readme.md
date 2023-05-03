# 高性能websocket推送组件

基本架构图

![image-20230423222249615](https://i.328888.xyz/2023/04/23/iS36no.jpeg)

## ws推送优化

1、写多读少的情况，使用epoll io多路复用避免创建读的协程。

2、分片存储连接，降低锁的粒度，提高推送的并发。

3、写的时候使用缓存，批量定时写入，减少系统调用、协程的调度。



### 1、epoll

1、实现参考，https://github.com/eranyanay/1m-go-websockets   只适用在写多读少的情况下，并且读之后的业务耗时的操作不要太频繁。

 这里读的时候要考虑到半消息的情况，https://studygolang.com/topics/13377  当收到websocket数据包不全的情况下，go net官方库当不可读的时候是会阻塞的。如果这里如果我们也阻塞的的话，那就要等他读完我们才能处理其他的。

**解决方法**：如果如果没有读完的，就缓存起来，读的起始位置就回退到上次读出完整数据的位置，这里需要自己diy bufio.Reader 缓存没有读完的消息。

### 2、使用写缓存

如果在小包特别多得情况下，直接发送会有很多系统调用,gmp模型，每一次系统调用go都会有一次调度 会消耗cpu。在实时性要求不高的情况下，我们可以批量定时写入。

得益于github.com/gobwas/ws 提供的非常灵活的用法，我们可以暂时将我们要写的数据转为websocket数据包缓存起来， 批量定时写入。此外提前encode websocket数据包，也可以避免每一个连接都去encode 一遍websocket数据包,提高效率。

### 3、分片存储连接

主要是参考https://github.com/owenliang/go-push 

主要的思路就是分片存储连接，减少锁的粒度和推送的并发。

## 实现功能

推送数据到公共频道

推送数据到所有连接

推送数据到私有频道

## 订阅操作

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

## 后台推送

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

## 代码

具体代码仓库地址 https://github.com/luxun9527/gpush    练习时长还没到两年半，还有很多不完善的地方主要是提供一个实现的思路 如果对您有帮助，请帮我点一下star。

