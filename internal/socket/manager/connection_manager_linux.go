package manager

import (
	"github.com/luxun9527/gpush/internal/socket/global"
	"hash/fnv"
)

type PushType uint8

const (
	PushAll PushType = iota + 1
	PushRoom
	PushPerson
)

var CM *ConnectionManager

type PushJob struct {
	PushType
	roomID string
	data   []byte
	uid    string
}

// 选择一个合适的通道。
func (pushJob PushJob) getCid() int32 {
	var s uint32
	switch pushJob.PushType {
	case PushAll:
		//推送给所有用户的数据，要保证用户收到的数据是有序的只有选择同一个chan
	case PushRoom:
		//不同的room并发推送,但是推送到同一room中的数据有序的

		h := fnv.New32a()
		h.Write([]byte(pushJob.roomID))
		s = h.Sum32()
	case PushPerson:
		//不同的用户并发推送，但是用户收到的数据是有序的。
		h := fnv.New32a()
		h.Write([]byte(pushJob.uid))
		s = h.Sum32()
	}
	return int32(s % 20)
}

type ConnectionManager struct {
	buckets      []*Bucket
	dispatchChan chan *PushJob // 待分发消息队列
	Epoller      []*Epoller
}

func GetConnectionInfo() int {
	var count int
	for _, v := range CM.buckets {
		count += v.Count()
	}
	return count
}

// InitConnectionManager 初始化连接管理
func InitConnectionManager() {
	CM = &ConnectionManager{
		buckets:      make([]*Bucket, global.Config.Bucket.BucketCount),
		dispatchChan: make(chan *PushJob, global.Config.Bucket.DispatchChanSize),
	}
	for i, _ := range CM.buckets {
		//初始化所有bucket
		CM.buckets[i] = NewBucket(int32(i))
	}
	epollers := make([]*Epoller, 4)
	for i := 0; i < 4; i++ {
		epollers[i] = NewEpoller()
	}
	CM.Epoller = epollers
	//分发数据到所有bucket中不同chan
	go CM.dispatchToBucket()

}

// 将连接移除事件监听中
func (c *ConnectionManager) removeEpollerConn(id int64) error {
	i := id % 4
	return c.Epoller[i].Remove(int(id))
}

// 将连接加入到事件监听中
func (c *ConnectionManager) addEpollerConn(id int64) error {
	i := id % 4
	return c.Epoller[i].Add(int(id))
}

// 分发数据到到所有bucket的jobchan中
func (c *ConnectionManager) dispatchToBucket() {

	for job := range c.dispatchChan {
		cid := job.getCid()
		for _, v := range c.buckets {
			v.messageChan[cid] <- job
		}
	}
}

// PushRoom 推送到room
func (c *ConnectionManager) PushRoom(room string, data []byte) {
	job := &PushJob{
		PushType: PushRoom,
		roomID:   room,
		data:     data,
	}

	c.dispatchChan <- job
}

// PushAll 推送给所有
func (c *ConnectionManager) PushAll(data []byte) {
	job := &PushJob{
		PushType: PushAll,
		data:     data,
	}
	c.dispatchChan <- job
}

// PushPerson 推送给所有
func (c *ConnectionManager) PushPerson(uid, roomID string, data []byte) {
	job := &PushJob{
		PushType: PushPerson,
		uid:      uid,
		data:     data,
		roomID:   roomID,
	}
	c.dispatchChan <- job
}

// AddConnection 添加连接
func (c *ConnectionManager) AddConnection(conn *Connection) {
	b := c.getBucket(conn.ID)
	b.AddConn(conn)
	global.Prometheus.WsConnections.WithLabelValues().Inc()
	global.Prometheus.WsTotalConnections.WithLabelValues().Inc()
}

// NotifyRead 通知连接读取
func (c *ConnectionManager) NotifyRead(fd int) {
	b := c.getBucket(int64(fd))
	i := fd % 10
	b.notify[i] <- fd
}

// CloseConnection 关闭连接
func (c *ConnectionManager) CloseConnection(fd int) {
	b := c.getBucket(int64(fd))
	conn, ok := b.GetConnection(int64(fd))
	if ok {
		conn.Close()
	}
}

// DelConnection 从存储所有连接中删除连接
func (c *ConnectionManager) DelConnection(conn *Connection) {
	b := c.getBucket(conn.ID)
	b.delConn(conn)
}

// JoinPublicRoom 加入共有的房间
func (c *ConnectionManager) JoinPublicRoom(roomID string, conn *Connection) {
	if ok := conn.isSubbed(roomID); ok {
		return
	}
	//新增连接上的room
	conn.subRoom(roomID, Public)
	b := c.getBucket(conn.ID)
	b.joinPublicRoom(roomID, conn)
}

// LeavePublicRoom 离开公有的房间
func (c *ConnectionManager) LeavePublicRoom(roomID string, conn *Connection) {
	//删除连接上的room
	conn.UnSubRoom(roomID)
	c.leavePublicRoom(roomID, conn)
}
func (c *ConnectionManager) leavePublicRoom(roomID string, conn *Connection) {
	//删除连接上的room
	b := c.getBucket(conn.ID)
	b.leavePublicRoom(roomID, conn)
}

// JoinPrivateRoom 加入私有的房间
func (c *ConnectionManager) JoinPrivateRoom(roomID string, conn *Connection) {
	//删除连接上的room
	ok := conn.isSubbed(roomID)
	if ok {
		return
	}
	conn.subRoom(roomID, Private)

}

// LeavePrivateRoom 离开私有房间
func (c *ConnectionManager) LeavePrivateRoom(roomID string, conn *Connection) {
	//删除连接上的room
	conn.UnSubRoom(roomID)
}
func (c *ConnectionManager) getBucket(connID int64) *Bucket {
	return c.buckets[connID%int64(len(c.buckets))]
}

// LevelAll 离开所有保存了连接的地方,当连接退出关闭的时候
func (c *ConnectionManager) LevelAll(conn *Connection) {
	c.removeEpollerConn(conn.ID)
	c.DelConnection(conn)
	conn.lock.RLock()
	defer conn.lock.RUnlock()
	for room, roomType := range conn.subbedRooms {
		if roomType == Public {
			c.leavePublicRoom(room, conn)
		}
	}
}
func (c *ConnectionManager) Login(conn *Connection) {
	b := c.getBucket(conn.ID)
	b.addLoggedConnection(conn)

}

func (c *ConnectionManager) Logout(conn *Connection) {
	//连接上订阅私有room
	publicRoom := make(map[string]RoomType, len(conn.subbedRooms))
	conn.lock.Lock()
	for room, roomType := range conn.subbedRooms {
		if roomType == Public {
			publicRoom[room] = Public
		}
	}
	conn.subbedRooms = publicRoom
	conn.lock.Unlock()
	b := c.getBucket(conn.ID)
	b.deleteLoggedConnection(conn)

}
