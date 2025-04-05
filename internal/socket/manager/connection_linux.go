package manager

import (
	"bufio"
	"errors"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/luxun9527/gpush/internal/socket/global"
	"github.com/luxun9527/gpush/tools"
	"github.com/luxun9527/zlog"
	"github.com/spf13/cast"
	"go.uber.org/atomic"
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
	isClosed      bool
	ip            string
	handlerFunc   HandlerFunc
	writeRate     *time.Ticker
	writeBuf      *bufio.Writer
	closeFunc     sync.Once
	epollerReader *tools.Reader
	readBuf       *bufio.Reader
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

func (conn *Connection) Send(data []byte) {
	select {
	case conn.write <- data:
	default:
	}

}

func (conn *Connection) Read(data []byte) (n int, err error) {
	//直接系统调用读取
	n, err = syscall.Read(int(conn.ID), data)
	if err != nil {
		if errors.Is(err, syscall.EAGAIN) {
			return 0, io.EOF
		}
		return 0, err
	}
	return n, nil
}

func (conn *Connection) ReadMessage() {
	defer conn.Close()
	for {
		frame, err := ws.ReadFrame(conn.readBuf)
		if err != nil {
			zlog.Debugf("read error %v", err)
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
				zlog.Debugf("decompressFrame failed %v", err)
				return
			}
		}
		conn.handlerFunc(frame.Payload, conn)

	}

}

func (conn *Connection) EpollerReadMessage() {
	for {
		frame, err := ws.ReadFrame(conn.epollerReader)
		if err != nil {
			//处理半消息的情况，正常情况 ws.ReadFrame是会阻塞等待读取出完整的一条ws消息，但是我们使用了epoll,不是所有的连接都分配了一个协程，不能让其阻塞，否则会阻塞协程后续数据的处理
			//所以这里需要处理一次没读取完整的情况，缓存到内存下次再读。
			if errors.Is(err, io.ErrUnexpectedEOF) {
				//如果是没有读完，回退到上一次读出完整数据的位置.
				conn.epollerReader.Rewind()
			}
			zlog.Debugf("read error %v", err)
			break
		}
		//更新读出上一条完整数据的位置。
		conn.epollerReader.UpdateLastMessagePos()
		if frame.Header.OpCode == ws.OpClose {
			conn.Close()
			return
		}
		frame = ws.UnmaskFrameInPlace(frame)
		if global.Config.Connection.IsCompress {
			frame, err = wsflate.DecompressFrame(frame)
			if err != nil {
				zlog.Debugf("decompressFrame failed %v", err)
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
		lastHeartbeat: time.Now(),
	}
	CM.AddConnection(nc)
	nc.writeRate = time.NewTicker(time.Millisecond * 10000)

	if global.Config.Connection.EnableWriteBuffer {
		nc.writeBuf = bufio.NewWriterSize(conn, global.Config.Connection.WriteBuf)
		nc.writeRate = time.NewTicker(time.Millisecond * time.Duration(global.Config.Connection.WriteRate))
	}

	if global.Config.Connection.EableEpoller {
		nc.epollerReader = tools.NewReaderSize(nc, global.Config.Connection.ReadBuf)
		if err := CM.addEpollerConn(ID); err != nil {
			zlog.Errorf("add conn to epoller failed %v", err)
			return nil, err
		}

	} else {
		nc.readBuf = bufio.NewReaderSize(conn, global.Config.Connection.ReadBuf)
		go nc.ReadMessage()
	}

	go nc.WriteLoop()
	return nc, nil
}

func (conn *Connection) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			zlog.Debugf("recover from read %v", err)
		}
		conn.Close()
	}()
	if global.Config.Connection.EnableWriteBuffer {
		for {
			select {
			case data := <-conn.write:
				global.Prometheus.WsBytesSent.WithLabelValues().Add(cast.ToFloat64(len(data)))
				if _, err := conn.writeBuf.Write(data); err != nil {
					zlog.Debugf("Write from read %v", err)
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
	} else {

		for {
			select {
			case data := <-conn.write:
				global.Prometheus.WsBytesSent.WithLabelValues().Add(cast.ToFloat64(len(data)))

				if _, err := conn.Conn.Write(data); err != nil {
					zlog.Debugf("writer data failed %v", err)
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
			}
		}

	}

}

func (conn *Connection) Close() {

	conn.closeFunc.Do(func() {
		global.Prometheus.WsConnections.WithLabelValues().Dec()
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
	conn.lastHeartbeat = time.Now().Add(time.Millisecond * time.Duration(global.Config.Connection.TimeOut))
}
