package test

import (
	"context"
	pb "github.com/luxun9527/gpush/proto"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"
)

// go test -v push_test.go -test.run TestPush
func TestPush(t *testing.T) {
	conn, err := grpc.Dial("192.168.2.159:10067", grpc.WithBlock(), grpc.WithInsecure())
	//	conn, err := grpc.Dial("127.0.0.1:10067", grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Println(err)
		return
	}
	var c atomic.Int64
	client := pb.NewProxyClient(conn)

	data := []byte{ //'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
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

	for j := 1; j <= 500000; j++ {
		client.PushData(context.Background(), &pb.Data{
			Uid:   "",
			Topic: "test",
			Data:  data,
		})

		c.Store(c.Inc())
	}

	time.Sleep(time.Second * 10)

}

func TestServer(t *testing.T) {

}

func TestClient(t *testing.T) {

}
