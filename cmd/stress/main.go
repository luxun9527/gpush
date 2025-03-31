package main

import (
	"flag"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

var (
	currentReceived atomic.Int64
	lastTime        int64
	lastReceived    int64
)

func main() {
	var url string
	// /root/ws/cmd/stress/stress --count=5000 --url=ws://192.168.2.99:9992/ws
	var count int64

	flag.StringVar(&url, "url", "ws://192.168.2.159:9995/ws", "url")
	flag.Int64Var(&count, "count", 5000, "count")
	flag.Parse()
	for i := int64(0); i < count; i++ {
		time.Sleep(time.Microsecond)
		go connect(url)
	}
	go func() {
		for {
			time.Sleep(time.Second)
			now := time.Now().UnixMilli()
			cr := currentReceived.Load()
			s := now - lastTime
			receivedLastSecond := (cr - lastReceived) / s
			lastReceived = cr
			lastTime = now
			log.Printf("当前收到 %v条 平均每毫秒收到 %v条\n", cr, receivedLastSecond)
		}

	}()
	time.Sleep(time.Hour)
}
func connect(url string) {
	c := websocket.DefaultDialer
	conn, _, err := c.Dial(url, nil)
	if err != nil {
		log.Printf("connect failed err = %v", err.Error())
		return
	}
	ReadMessage(conn)
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
		dataType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("data %v err %v %v", string(data), err, dataType)
			return
		}
		currentReceived.Inc()

	}
}

//func SendMessage(conn *websocket.Conn) {
//	for i := 0; i < 1000; i++ {
//		time.Sleep(time.Millisecond * 100)
//		hasSend.Inc()
//		conn.WriteMessage(websocket.TextMessage, []byte("testaestlasjdfljfalkjsdlfjalsdjfl"))
//	}
//	//for {
//	//	_, p, err := conn.ReadMessage()
//	//	if err != nil {
//	//		log.Println(err)
//	//		return
//	//	}
//	//	log.Println(string(p))
//	//}
//
//}
