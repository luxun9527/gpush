package manager

import (
	"bufio"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/tools"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"syscall"
	"time"
)

type RoomType uint8

const (
	Private RoomType = iota + 1
	Public
)

type HandlerFunc func([]byte, *Connection)
type Connection struct {
	net.Conn
	//连接的fd
	ID          int64
	write       chan []byte
	lock        sync.RWMutex
	subbedRooms map[string]RoomType
	//是否关闭
	isClosed    bool
	ip          string
	handlerFunc HandlerFunc
	writeRate   *time.Ticker
	writeBuf    *bufio.Writer
	closeFunc   sync.Once
	//读数据到这个buf中
	readBuf       *tools.Reader
	lastHeartbeat time.Time
	uid           atomic.String
}

// SetUid 设置uid
func (conn *Connection) SetUid(uid string) {
	conn.uid.Store(uid)
}

// GetUid 获取UID
func (conn *Connection) GetUid() string {
	return conn.uid.Load()
}

// IsLogin 是否登录
func (conn *Connection) IsLogin() bool {
	return conn.GetUid() != ""
}

// Send 发送数据 看你的选择，如果消息很重要一定要推出去，当不可写的时候你应该要关闭连接
func (conn *Connection) Send(data []byte) {
	select {
	case conn.write <- data:
	default:
	}

}

func (conn *Connection) Read(data []byte) (n int, err error) {
	n, err = syscall.Read(int(conn.ID), data)
	if err != nil {
		if err == syscall.EAGAIN {
			return 0, io.EOF
		}
		return 0, err
	}
	return n, nil
}

func (conn *Connection) ReadMessage() {

	for {
		frame, err := ws.ReadFrame(conn.readBuf)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				//如果是没有读完，回退到上一次读出完整数据的位置.
				conn.readBuf.GoBack()
			}
			break
		}
		//更新读出上一条完整数据的位置。
		conn.readBuf.UpdateLastMessagePos()
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
		conn.handlerFunc(frame.Payload, conn)

	}

}

func NewConnection(conn net.Conn, f HandlerFunc) (*Connection, error) {
	ID := tools.WebsocketFD(conn)

	nc := &Connection{
		Conn:          conn,
		ip:            conn.RemoteAddr().String(),
		ID:            ID,
		write:         make(chan []byte, 50),
		subbedRooms:   make(map[string]RoomType, 5),
		handlerFunc:   f,
		writeRate:     time.NewTicker(time.Millisecond * time.Duration(global.Config.Connection.WriteRate)),
		lastHeartbeat: time.Now(),
		writeBuf:      bufio.NewWriterSize(conn, global.Config.Connection.WriteBuf),
	}
	nc.readBuf = tools.NewReaderSize(nc, global.Config.Connection.ReadBuf)
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
		CM.LevelAll(conn)
		conn.Conn.Close()
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
func (conn *Connection) subRoom(roomID string, roomType RoomType) {
	conn.lock.Lock()
	conn.subbedRooms[roomID] = roomType
	conn.lock.Unlock()
}
func (conn *Connection) UnSubRoom(roomID string) {
	conn.lock.Lock()
	delete(conn.subbedRooms, roomID)
	conn.lock.Unlock()
}
func (conn *Connection) unSubAll(roomID string) {
	conn.lock.Lock()
	delete(conn.subbedRooms, roomID)
	conn.lock.Unlock()
}

func (conn *Connection) KeepAlive() {
	conn.lastHeartbeat = time.Now().Add(time.Millisecond * time.Duration(global.Config.Connection.TimeOut))
}
