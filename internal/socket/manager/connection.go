package manager

import (
	"bufio"
	"bytes"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/mofei1/gpush/internal/socket/global"
	"github.com/mofei1/gpush/tools"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

var (
	Received     atomic.Int64
	Ship         atomic.Int64
	LastReceived atomic.Int64
	SendCount    atomic.Int64
)

type HandlerFunc func([]byte, *Connection)
type Connection struct {
	net.Conn
	ID          int64
	write       chan []byte
	lock        sync.RWMutex
	subbedRooms map[string]struct{}
	isClosed    bool
	ip          string
	handlerFunc HandlerFunc
	//写频率
	writeRate *time.Ticker
	writeBuf  *bufio.Writer
	closeFunc sync.Once
	//读数据到这个buf中
	readBuf       *bytes.Buffer
	lastHeartbeat time.Time
	isLogin       atomic.Bool
	Uid           string
}

func (conn *Connection) SetLoginStatus(status bool) {
	conn.isLogin.Store(status)
}
func (conn *Connection) IsLogin() bool {
	return conn.isLogin.Load()
}
func (conn *Connection) Send(data []byte) {
	//SendCount.Inc()
	select {
	case conn.write <- data:
	default:
		//Ship.Inc()
	}

}

func (conn *Connection) Read(fd int) {
	for {
		buf := make([]byte, 100)
		//考虑一次读多条和一次读不完一条的情况。
		n, err := syscall.Read(fd, buf)
		if err != nil {
			//当读出错就返回
			return
		}
		//todo 自己实现解码websocket数据包，写到缓存并不是最好的选择。
		conn.readBuf.Write(buf[:n])
		//循环读
		for {
			frame, err := ws.ReadFrame(conn.readBuf)
			if err != nil {
				if err == io.ErrUnexpectedEOF {
					conn.readBuf.Write(buf[:n])
				}
				break
			}
			if frame.Header.OpCode == ws.OpClose {
				conn.Close()
				return
			}
			frame = ws.UnmaskFrameInPlace(frame)
			if global.Config.Connection.IsCompress {
				frame, err = wsflate.DecompressFrame(frame)
				if err != nil {
					return
				}
			}
			//Received.Inc()
			conn.handlerFunc(frame.Payload, conn)

		}

	}
}

func NewConnection(conn net.Conn, f HandlerFunc) (*Connection, error) {
	ID := tools.WebsocketFD(conn)

	nc := &Connection{
		Conn:          conn,
		ip:            conn.RemoteAddr().String(),
		ID:            ID,
		write:         make(chan []byte, 50),
		subbedRooms:   make(map[string]struct{}, 5),
		handlerFunc:   f,
		writeRate:     time.NewTicker(time.Millisecond * time.Duration(global.Config.Connection.WriteRate)),
		readBuf:       bytes.NewBuffer(make([]byte, 0, 100)),
		lastHeartbeat: time.Now(),
	}
	//nc.writeBuf = bufio.NewWriterSize(nc, global.Config.Connection.WriteBuf)
	nc.writeBuf = bufio.NewWriterSize(conn, global.Config.Connection.WriteBuf)
	CM.AddConnection(nc)
	if err := CM.addEpollerConn(ID); err != nil {
		global.L.Error("add conn to epoller failed", zap.Error(err))
		return nil, err
	}

	go nc.WriteLoop()
	return nc, nil
}

func (conn *Connection) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			global.L.Debug("recover from read", zap.Any("err", err))
		}
		conn.Close()
	}()
	for {
		select {
		case data := <-conn.write:
			if _, err := conn.writeBuf.Write(data); err != nil {
				return
			}
		case <-conn.writeRate.C:
			//如果关闭就返回
			if conn.isClosed {
				return
			}
			//心跳超时
			if conn.lastHeartbeat.Add(time.Millisecond * time.Duration(global.Config.Connection.TimeOut)).Before(time.Now()) {
				return
			}
			//写到连接中
			if conn.writeBuf.Available() > 0 {
				if err := conn.writeBuf.Flush(); err != nil {
					return
				}
			}

		}
	}
}

//func (conn *Connection) Write(data []byte) (int, error) {
//	var nn int
//	for {
//		n, err := syscall.Write(int(conn.ID), data[nn:])
//		if n > 0 {
//			nn += n
//		}
//		if nn == len(data) {
//			return nn, err
//		}
//		if err != nil {
//			return 0, err
//		}
//		if n == 0 {
//			return nn, io.ErrUnexpectedEOF
//		}
//	}
//}
func (conn *Connection) Close() {

	conn.closeFunc.Do(func() {
		CM.removeEpollerConn(conn.ID)
		conn.Conn.Close()
		conn.levelAll()
		conn.isClosed = true
	})

}

//是否订阅某个room
func (conn *Connection) isSubbed(roomID string) bool {
	conn.lock.RLock()
	_, ok := conn.subbedRooms[roomID]
	conn.lock.RUnlock()
	return ok
}
func (conn *Connection) subRoom(roomID string) {
	conn.lock.Lock()
	conn.subbedRooms[roomID] = struct{}{}
	conn.lock.Unlock()
}
func (conn *Connection) unSubRoom(roomID string) {
	conn.lock.Lock()
	delete(conn.subbedRooms, roomID)
	conn.lock.Unlock()
}

// LevelAll 只有断开才会执行
func (conn *Connection) levelAll() {
	CM.DelConnection(conn)
	conn.lock.RLock()
	defer conn.lock.RUnlock()
	for room, _ := range conn.subbedRooms {
		CM.LeaveRoom(room, conn)
	}
}
func (conn *Connection) KeepAlive() {
	conn.lastHeartbeat = time.Now().Add(time.Millisecond * time.Duration(global.Config.Connection.TimeOut))
}
