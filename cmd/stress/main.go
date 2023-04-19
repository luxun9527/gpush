package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"log"
	"time"
)

var currentReceived, lastReceived atomic.Int64
var hasSend atomic.Int64

func main() {
	var url string
	// /root/ws/cmd/stress/stress --count=5000 --url=ws://192.168.2.99:9992/ws
	var count int64

	flag.StringVar(&url, "url", "ws://192.168.2.99:9992/ws", "url")
	flag.Int64Var(&count, "count", 5000, "count")
	flag.Parse()
	for i := int64(0); i < count; i++ {
		time.Sleep(time.Millisecond * 2)
		go connect(url)
	}
	go func() {
		for {
			time.Sleep(time.Second * 1)
			log.Printf("发送了多少条 %v\n", hasSend.Load())
		}

	}()
	time.Sleep(time.Hour)
}
func connect(url string) {

	c := websocket.DefaultDialer
	//c.EnableCompression = true
	conn, _, err := c.Dial(url, nil)
	if err != nil {
		log.Printf("connect failed err = %v", err.Error())
		return
	}
	SendMessage(conn)
}
func ReadMessage(conn *websocket.Conn) {
	//conn.EnableWriteCompression(true)
	//conn.SetCompressionLevel(9)
	conn.WriteJSON(map[string]interface{}{"code": 1, "topic": "test"})
	//go func() {
	//
	//	for {
	//		time.Sleep(time.Second * 10)
	//		conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	//
	//	}
	//}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("err", err)
			return
		}

		currentReceived.Inc()

	}
}
func SendMessage(conn *websocket.Conn) {
	for i := 0; i < 1000; i++ {
		time.Sleep(time.Millisecond * 100)
		hasSend.Inc()
		conn.WriteMessage(websocket.TextMessage, []byte("testaestlasjdfljfalkjsdlfjalsdjfl"))
	}
	//for {
	//	_, p, err := conn.ReadMessage()
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	log.Println(string(p))
	//}

}
