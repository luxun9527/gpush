package manager

import (
	"bufio"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/tools"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net"
	"sync"
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
	readBuf *bufio.Reader
	uid     atomic.String
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

// Send
func (conn *Connection) Send(data []byte) {
	select {
	case conn.write <- data:
	default:
	}

}
func (conn *Connection) ReadMessage() {

	for {
		frame, err := ws.ReadFrame(conn.readBuf)
		if err != nil {
			global.L.Debug("read error", zap.Error(err))
			break
		}
		//更新读出上一条完整数据的位置。
		if frame.Header.OpCode == ws.OpClose {
			conn.Close()
			return
		}
		frame = ws.UnmaskFrameInPlace(frame)
		if global.Config.Connection.IsCompress {
			frame, err = wsflate.DecompressFrame(frame)
			if err != nil {
				global.L.Debug("decompress frame read error", zap.Error(err))
				return
			}
		}
		conn.handlerFunc(frame.Payload, conn)

	}

}
func NewConnection(conn net.Conn, f HandlerFunc) (*Connection, error) {
	ID := tools.WebsocketFD(conn)

	nc := &Connection{
		Conn:        conn,
		ip:          conn.RemoteAddr().String(),
		ID:          ID,
		write:       make(chan []byte, 50),
		subbedRooms: make(map[string]RoomType, 5),
		handlerFunc: f,
		writeRate:   time.NewTicker(time.Millisecond * time.Duration(global.Config.Connection.WriteRate)),
		writeBuf:    bufio.NewWriterSize(conn, global.Config.Connection.WriteBuf),
		readBuf:     bufio.NewReaderSize(conn, 4096),
	}
	CM.AddConnection(nc)
	go nc.ReadMessage()
	go nc.WriteLoop()
	return nc, nil
}

// compare
func (conn *Connection) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			global.L.Debug("recover from write", zap.Any("err", err))
		}
		conn.Close()
	}()

	for {
		select {
		case data := <-conn.write:
			if _, err := conn.writeBuf.Write(data); err != nil {
				global.L.Debug("write data error", zap.Error(err))
				return
			}
		case <-conn.writeRate.C:
			if err := conn.writeBuf.Flush(); err != nil {
				global.L.Debug("flush data error", zap.Error(err))
				return
			}
		}
	}

}

func (conn *Connection) Close() {

	conn.closeFunc.Do(func() {
		CM.LevelAll(conn)
		conn.Conn.Close()
		conn.isClosed = true
	})

}

// 是否订阅某个room
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
	conn.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
}
