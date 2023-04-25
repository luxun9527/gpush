package test

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	pb "github.com/luxun9527/gpush/proto"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"testing"
	"time"
)

var count atomic.Int64

func connect(i int) {
	// nohup /root/ws/cmd/stress/stress --url=ws://192.168.138.99:9992/ws --count=5000
	c := websocket.DefaultDialer
	conn, _, err := c.Dial("ws://192.168.153.99:9992/ws", nil)
	if err != nil {
		log.Printf("connect failed err = %v", err.Error())
		return
	}
	var sum int
	conn.WriteJSON(map[string]interface{}{"code": 1, "topic": "test"})
	go func() {
		for {
			time.Sleep(time.Second * 10)
			conn.WriteMessage(websocket.TextMessage, []byte("ping"))
		}
	}()
	for {
		_, data, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			continue
		}
		if string(data) != "pong" {
			log.Println(string(data), sum)
			sum++
		}

	}
}
func TestQps(t *testing.T) {

	connect(1)
}

//go test -v push_test.go -test.run TestPush
func TestPush(t *testing.T) {
	conn, err := grpc.Dial("192.168.2.99:10067", grpc.WithBlock(), grpc.WithInsecure())
	//	conn, err := grpc.Dial("127.0.0.1:10067", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return
	}
	var c atomic.Int64
	client := pb.NewProxyClient(conn)

	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		//'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	go func() {
		for {
			time.Sleep(time.Second)
			log.Println(c.Load())
		}

	}()

	for j := 1; j <= 1000000; j++ {
		client.PushData(context.Background(), &pb.Data{
			Uid:   "",
			Topic: "test",
			Data:  data,
		})

		c.Store(c.Inc())
	}

	time.Sleep(time.Second * 10)

}

type Conn struct {
	write chan []byte
	C     *websocket.Conn
}

func TestServer(t *testing.T) {
	r := gin.New()
	//var i int64
	var m []*Conn
	var sendCount atomic.Int64
	go func() {
		for {
			time.Sleep(time.Second)
			log.Println(sendCount.Load())
		}

	}()
	r.GET("/ws", func(c *gin.Context) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			//	global.L.Error("upgrade 连接失败",zap.Error(err))
			log.Println(err)
			return
		}
		co := &Conn{
			write: make(chan []byte, 200),
			C:     conn,
		}
		m = append(m, co)
		go func() {
			for {
				d := <-co.write
				co.C.WriteMessage(websocket.BinaryMessage, d)
				sendCount.Inc()
			}
		}()
	})
	go func() {
		d := []byte{'a'}
		time.Sleep(time.Second * 20)
		for {
			for _, c := range m {
				st := time.Now().UnixMilli()
				select {
				case c.write <- d:
				default:
					//log.Println("fulled")
				}
				et := time.Now().UnixMilli()
				if et-st > 2 {
					log.Println("cost", et-st)
				}
				//c.write<-[]byte{'a'}
			}

		}

	}()

	r.Run(":9998")
}

func TestClient(t *testing.T) {
	c := websocket.DefaultDialer
	conn, _, err := c.Dial("ws://192.168.138.99:9998/ws", nil)
	if err != nil {
		log.Printf("connect failed err = %v", err.Error())
		return
	}
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}
	}
}
func TestW(t *testing.T) {

}
