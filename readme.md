# 高性能websocket推送组件

![image-20230423222249615](https://i.328888.xyz/2023/04/23/iS36no.jpeg)

## ws推送优化

1、写多读少的情况，使用epoll 避免创建读的协程。

2、分片存储连接，降低锁的粒度，提高推送的并发。

3、写的时候使用缓存，批量定时写入，减少系统调用、协程的调度。



### 1、epoll

1、实现参考，https://github.com/eranyanay/1m-go-websockets   只适用在写多读少的情况下，并且读之后的业务耗时的操作不要太频繁。

https://github.com/eranyanay/1m-go-websockets/blob/master/3_optimize_ws_goroutines/server.go 这里读的时候要考虑到半消息的情况，https://studygolang.com/topics/13377  当收到websocket数据包不全的情况下，go net官方库当不可读的时候是会阻塞的。

我自己写的一个测试 github.com/gobwas/ws 提供了非常灵活的用法我们可以可以直接将一个websocket数据包转为字节。

```go
func main() { {
     dialer := gws.DefaultDialer
     conn, _, _, err := dialer.Dial(context.Background(), "ws://10.18.13.129:8989/ws")
     if err != nil {
        log.Fatalln("err", err)
    }

     message := NewMessage(gws.OpText, []byte("abcdefjhigklmnopqrusasdflsdfasdl"))
     b, _ := message.ToBytes()
     conn.Write(b[:5])
     log.Println("发送前5个字节")
     time.Sleep(time.Second * 3)
     conn.Write(b[5:])
     log.Println("发送后面的字节")

     select {}
 }

   type Message struct {
    messageType gws.OpCode
    data        []byte
}

func NewMessage(code gws.OpCode, data []byte) Message {
    return Message{
        messageType: code,
        data:        data,
    }
}
func (m Message) ToBytes()([]byte, error) {
	var res bytes.Buffer
	frame := gws.NewFrame(m.messageType, true, m.data)
	if err := gws.WriteFrame(&res, frame); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}
2023/04/18 18:09:55 发送前5个字节
2023/04/18 18:09:58 发送后面的字节
```



```go
func main() {
    engine := gin.New()
    engine.GET("/ws", Connect)
    engine.Run(":8989")
}
func Connect(c *gin.Context) {
    var httpUpgrade gws.HTTPUpgrader
    conn, _, _, err := httpUpgrade.Upgrade(c.Request, c.Writer)
    if err != nil {
       return
   }

    if err != nil {
       return
   }
     frame, err := gws.ReadFrame(conn)
       if err != nil {
          log.Println("err", err)
          return
      }
    log.Println(string(frame.Payload))
	//最终要等到数据全部到，go net官方库当不可读的时候是阻塞的。
   2023/04/18 18:09:58 abcdefjhigklmnopqrusasdflsdfasdl

}
```

**解决方法**：将数据读出来，如果io.ReadFull如果没有读完的，就回退到上次读出完整数据的位置，这里需要自己diy bufio.Reader ,当没有读出“半消息的时候缓存起来”。

### 2、使用写缓存

如果在小包特别多得情况下，直接发送会有很多系统调用,gmp模型，每一次系统调用go都会有一次调度 会消耗cpu。在实时性要求不高的情况下，我们可以批量定时写入。

得益于github.com/gobwas/ws 提供的非常灵活的用法，我们可以暂时将我们要写的数据转为websocket数据包缓存起来， 批量定时写入。此外提前encode websocket数据包，也可以避免每一个连接都去encode 一遍websocket数据包,提高效率。

```go
 type Message struct {
    messageType gws.OpCode
    data        []byte
}
func (m Message) ToBytes()([]byte, error) {
	var res bytes.Buffer
	frame := gws.NewFrame(m.messageType, true, m.data)
	if err := gws.WriteFrame(&res, frame); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}
for {
    select {
       case data := <-conn.write:
       if _, err := conn.writeBuf.Write(data); err != nil {
          return
      }
      //writeBuf 为*bufio.Writer。
       case <-conn.writeRate.C:
    
       //写到连接中
       if conn.writeBuf.Available() > 0 {
          if err := conn.writeBuf.Flush(); err != nil {
             return
         }
      }

   }
}
```

### 3、分片存储连接

主要是参考https://github.com/owenliang/go-push 

主要的思路就是分区map,连接分散到到不同的map中，提高锁的粒度和推送的并发读。同时如果是读多写少的情况可以使用sync.Map提高效率

## 实现功能

推送数据到公共频道

推送数据到所有连接

推送数据到私有频道

## 订阅操作

客户端请求数据的基本结构

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

