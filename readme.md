# ws推送组件

## 实现功能

### 推送数据到公共频道

### 推送数据到所有连接

### 推送数据到私有频道

## 订阅操作

客户端请求数据的基本结构

```json
{
    "code":1,
    "topic":"test",
    "data":object
}
```

code 表示操作的类型

- 1订阅公共频道。
- 2取消订阅公共频道。
- 3登录操作，需要你自己实现登录的逻辑。
- 4登出操作,需要你自己实现登出的逻辑，会退出你订阅的所有私有频道。
- 5订阅私有频道，需要登录。
- 6取消订阅私有频道，需要登录。

topic 你订阅的频道的名称

data object类型，登录的时候你可以使用这个字段传你的token

## 后台推送

后台推送提供http接口和grpc接口

```json
{
    "uid":"123456",
    "topic":"test",
    "data":[97,98,99]
}
```

uid    :  用户id当推送给私有频道指定用户的id，否则为空

topic : 频道的名称，当你想推送给所有的时候为空

data : 具体的推送的内容

## ws推送优化

1、写多读少的情况，使用epoll 避免创建读的协程。

2、分片存储连接，降低锁的粒度，提高推送的并发。

3、写的时候使用缓存，批量定时写入，减少系统调用、协程的调度。

4、直接使用tcp连接升级。

5、零拷贝。



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

解决方法，使用syscall.Read，将websocket数据包读出后再处理，而不是阻塞。

```go
for {
    buf := make([]byte, 100)
    //考虑一次读多条和一次读不完一条的情况。
    n, err := syscall.Read(fd, buf)
    if err != nil {
       //当读出错就返回
       return
   }
    //写到缓存区中
    conn.readBuf.Write(buf[:n])
    //循环读
    for {
        //todo 这个地方可以优化，自己实现websocket包的解码，长度不够就写回去不是最佳选择。
       frame, err := ws.ReadFrame(conn.readBuf)
       if err != nil {
          if err == io.ErrUnexpectedEOF {
             //长度没达到就写回去，而不是阻塞
             conn.readBuf.Write(buf[:n])
         }
          break
      }
       if frame.Header.OpCode == ws.OpClose {
          conn.Close()
          return
      }
       frame = ws.UnmaskFrameInPlace(frame)
     
       log.Println(string(frame.Payload))
   }

}
```

### 2、使用写缓存

如果在小包特别多得情况下，直接发送会有很多系统调用,gmp模型，每一次系统调用go都会有一次调度 会消耗cpu。在实时性要求不高的情况下，我们可以批量定时写入。

得益于github.com/gobwas/ws 提供的非常灵活的用法，我们可以暂时将我们要写的数据转为websocket数据包缓存起来。批量定时写入

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

### 4、零拷贝

在大包多的情况是使用零拷贝，好像也会提高性能。

 https://zhuanlan.zhihu.com/p/308054212

https://mp.weixin.qq.com/s/wSaJYg-HqnYY4SdLA2Zzaw

### 5、基于tcp升级

github.com/gobwas/ws 也有直接基于tcp升级的用法，避免http分配的4k读缓存。



### 6、具体代码

具体代码仓库地址 https://github.com/mofei1/gpush    练习时长还没到两年半，还有很多不完善的地方主要是提供一个实现的思路 如果对您有帮助，请帮我点一下star。

