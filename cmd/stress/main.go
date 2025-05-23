package main

import (
	"flag"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

var (
	currentReceived      atomic.Int64
	currentReceivedBytes atomic.Int64
	lastTime             int64
	lastReceived         int64
	lastReceivedBytes    int64
)
var (
	enableCompress bool
	count          int64
	url            string
)

func init() {
	flag.StringVar(&url, "url", "ws://192.168.2.159:9995/ws", "url")
	flag.Int64Var(&count, "count", 5000, "count")
	flag.BoolVar(&enableCompress, "ec", false, "是否开启压缩")
	flag.Parse()
}

func main() {

	// /root/ws/cmd/stress/stress --count=5000 --url=ws://192.168.2.99:9992/ws

	for i := int64(0); i < count; i++ {
		time.Sleep(time.Microsecond)
		go connect(url)
	}
	go func() {
		for {
			time.Sleep(time.Second)
			now := time.Now().Unix()
			cr := currentReceived.Load()
			crb := currentReceivedBytes.Load()
			s := now - lastTime
			receivedLastSecond := (cr - lastReceived) / s
			receivedBytesPerSec := crb - lastReceivedBytes
			lastReceived = cr
			lastTime = now
			lastReceivedBytes = crb
			log.Printf("当前收到 %v条 平均每秒收到 %v条 平均每秒收到%v字节数\n", cr, receivedLastSecond, receivedBytesPerSec)
		}

	}()
	time.Sleep(time.Hour)
}
func connect(url string) {
	dialer := websocket.Dialer{}
	if enableCompress {
		dialer.EnableCompression = true
	}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		log.Printf("connect failed err = %v", err.Error())
		return
	}
	ReadMessage(conn)
}
func ReadMessage(conn *websocket.Conn) {
	if enableCompress {
		conn.EnableWriteCompression(true)
		conn.SetCompressionLevel(9)
	}

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
		currentReceivedBytes.Add(int64(len(data)))

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
