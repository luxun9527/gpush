package manager

import (
	sll "github.com/emirpasic/gods/lists/singlylinkedlist"
	"sync"
)

type Bucket struct {
	id          int32
	id2Conn     map[int64]*Connection //map[连接ID]*Connection 连接列表(key=连接唯一ID)
	rooms       map[string]*Room      //map[string]*Room 房间列表 。
	messageChan []chan PushJob
	id2ConnLock sync.RWMutex
	roomsLock   sync.RWMutex
	notify      []chan int
	loggedLock  sync.RWMutex
	LoggedConn  map[string]*sll.List //一个用户多次登录。
}

func (b *Bucket) Count() int {
	b.id2ConnLock.RLock()
	count := len(b.id2Conn)
	b.id2ConnLock.RUnlock()
	return count
}

//处理收到的数据
func (b *Bucket) handleMessage(c chan PushJob) {
	for job := range c {
		switch job.PushType {
		case PushAll:
			b.pushAll(job)
		case PushRoom:
			b.pushRoom(job)
		default:
			b.pushPerson(job)
		}
	}
}

// AddLoggedConnection 添加登录的连接
func (b *Bucket) addLoggedConnection(conn *Connection) {
	b.loggedLock.Lock()
	defer b.loggedLock.Unlock()
	c, ok := b.LoggedConn[conn.Uid]
	if !ok {
		c = sll.New()
		b.LoggedConn[conn.Uid] = c
	}
	c.Add(conn)
}

// DeleteLoggedConnection 删除登录的连接
func (b *Bucket) deleteLoggedConnection(conn *Connection) {
	b.loggedLock.Lock()
	defer b.loggedLock.Unlock()
	c, ok := b.LoggedConn[conn.Uid]
	if ok {
		if i := c.IndexOf(conn); i != -1 {
			c.Remove(i)
		}
		if c.Size() == 0 {
			delete(b.LoggedConn, conn.Uid)
		}
	}

}
func NewBucket(id int32) *Bucket {

	bucket := &Bucket{
		id:         id,
		id2Conn:    make(map[int64]*Connection, 200),
		rooms:      make(map[string]*Room, 100),
		LoggedConn: make(map[string]*sll.List, 1000),
	}
	messageChan := make([]chan PushJob, 20)
	cs := make([]chan int, 10)
	for i := 0; i < len(cs); i++ {
		cs[i] = make(chan int, 10)
		go bucket.readMessage(cs[i])
	}
	bucket.notify = cs
	for i := 0; i < len(messageChan); i++ {
		messageChan[i] = make(chan PushJob, 1000)
		go bucket.handleMessage(messageChan[i])
	}
	bucket.messageChan = messageChan
	return bucket
}

//读取通知读的操作。
func (b *Bucket) readMessage(fds chan int) {
	for fd := range fds {
		b.id2ConnLock.RLock()
		conn, ok := b.id2Conn[int64(fd)]
		b.id2ConnLock.RUnlock()
		if !ok {
			continue
		}
		conn.Read(fd)
	}
}

// AddConn 添加连接
func (b *Bucket) AddConn(conn *Connection) {
	b.id2ConnLock.Lock()
	b.id2Conn[conn.ID] = conn
	b.id2ConnLock.Unlock()
}

// DelConn 删除连接
func (b *Bucket) delConn(conn *Connection) {
	b.id2ConnLock.Lock()
	delete(b.id2Conn, conn.ID)
	b.id2ConnLock.Unlock()
}

// joinPublicRoom 加入共有的房间
func (b *Bucket) joinPublicRoom(roomID string, conn *Connection) {
	b.roomsLock.Lock()
	room, ok := b.rooms[roomID]
	if !ok {
		room = NewRoom(roomID)
		b.rooms[roomID] = room
	}
	b.roomsLock.Unlock()
	room.Join(conn)
}

// GetConnection 获取连接
func (b *Bucket) GetConnection(id int64) (*Connection, bool) {
	b.id2ConnLock.RLock()
	defer b.id2ConnLock.RUnlock()
	conn, ok := b.id2Conn[id]
	return conn, ok
}

// LeaveRoom 离开room
func (b *Bucket) leavePublicRoom(roomID string, conn *Connection) {
	b.roomsLock.RLock()
	room, ok := b.rooms[roomID]
	b.roomsLock.RUnlock()
	if !ok {
		return
	}
	room.Leave(conn)
	if room.Count() == 0 {
		b.roomsLock.Lock()
		delete(b.rooms, roomID)
		b.roomsLock.Unlock()
	}
}

//pushAll 推送给所有用户
func (b *Bucket) pushAll(job PushJob) {
	b.id2ConnLock.RLock()
	defer b.id2ConnLock.RUnlock()
	for _, conn := range b.id2Conn {
		conn.Send(job.data)
	}

}

//pushRoom 推送给指定的订阅
func (b *Bucket) pushRoom(job PushJob) {
	b.roomsLock.RLock()
	room, ok := b.rooms[job.roomID]
	b.roomsLock.RUnlock()
	if !ok {
		return
	}
	room.Push(job.data)
}

//pushPerson 推送给个人
func (b *Bucket) pushPerson(job PushJob) {
	// 这个锁的粒度比较大。
	b.loggedLock.RLock()
	defer b.loggedLock.RUnlock()
	lc, ok := b.LoggedConn[job.uid]
	if !ok {
		return
	}
	lc.Each(func(index int, value interface{}) {
		c := value.(*Connection)
		if ok := c.isSubbed(job.roomID); ok {
			c.Send(job.data)
		}
	})
}
