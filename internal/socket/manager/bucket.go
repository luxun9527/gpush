package manager

import (
	"sync"
)

type concurrenceList struct {
}
type Bucket struct {
	id          int32
	id2Conn     map[int64]*Connection //map[连接ID]*Connection 连接列表(key=连接唯一ID)
	rooms       map[string]*Room      //map[string]*Room 房间列表 。
	messageChan []chan PushJob
	id2ConnLock sync.RWMutex
	roomsLock   sync.RWMutex
	notify      []chan int
	loggedLock  sync.RWMutex
	Logged      map[string]*sync.Map //存储登录的连接,考虑到一个用户不同连接登录多个，很小的概率两个连接分到同一个bucket中。
}

func (b *Bucket) Count() int {
	b.id2ConnLock.RLock()
	count := len(b.id2Conn)
	defer b.id2ConnLock.RUnlock()
	return count
}
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
func (b *Bucket) AddLoginConnection(conn *Connection) {
	b.loggedLock.RLock()
	c, ok := b.Logged[conn.Uid]
	b.loggedLock.RUnlock()
	if !ok {
		c = &sync.Map{}
		b.loggedLock.Lock()
		b.Logged[conn.Uid] = c
		b.loggedLock.Unlock()
	}
	c.Store(conn.ID, conn)
}
func (b *Bucket) DeleteLoginConnection(conn *Connection) {
	b.loggedLock.RLock()
	c, ok := b.Logged[conn.Uid]
	b.loggedLock.RUnlock()
	if !ok {
		return
	}
	c.Delete(conn.ID)
	var i int32
	c.Range(func(_, _ any) bool {
		i++
		return true
	})
	if i == 0 {
		b.loggedLock.Lock()
		delete(b.Logged, conn.Uid)
		b.loggedLock.Unlock()
	}
}
func NewBucket(id int32) *Bucket {

	bucket := &Bucket{
		id:      id,
		id2Conn: make(map[int64]*Connection, 200),
		rooms:   make(map[string]*Room, 100),
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
func (b *Bucket) AddConn(conn *Connection) {
	b.id2ConnLock.Lock()
	b.id2Conn[conn.ID] = conn
	b.id2ConnLock.Unlock()

}
func (b *Bucket) DelConn(conn *Connection) {
	b.id2ConnLock.Lock()
	delete(b.id2Conn, conn.ID)
	b.id2ConnLock.Unlock()
}
func (b *Bucket) JoinRoom(roomID string, conn *Connection) {
	b.roomsLock.Lock()
	room, ok := b.rooms[roomID]
	if !ok {
		room = NewRoom(roomID)
		b.rooms[roomID] = room
	}
	b.roomsLock.Unlock()
	room.Join(conn)
}
func (b *Bucket) GetConnection(id int64) (*Connection, bool) {
	b.id2ConnLock.RLock()
	defer b.id2ConnLock.RUnlock()
	conn, ok := b.id2Conn[id]
	return conn, ok
}
func (b *Bucket) LeaveRoom(roomID string, conn *Connection) {
	b.roomsLock.RLock()
	room, ok := b.rooms[roomID]
	b.roomsLock.RUnlock()
	if !ok {
		return
	}
	if room.Count() == 0 {
		b.roomsLock.Lock()
		delete(b.rooms, roomID)
		b.roomsLock.Unlock()
	}
	room.Leave(conn)
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

	b.loggedLock.RLock()
	lc, ok := b.Logged[job.uid]
	b.loggedLock.RUnlock()
	if !ok {
		return
	}
	lc.Range(func(key, value any) bool {
		c := value.(*Connection)
		if ok := c.isSubbed(job.roomID); ok {
			c.Send(job.data)
		}
		return true
	})

}
