package manager

import (
	"github.com/mofei1/gpush/internal/socket/global"
)

type PushType int32

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

func (pushJob PushJob) GetRoomID() int32 {
	r := []byte(pushJob.roomID)
	var s int32
	for _, v := range r {
		s += int32(v)
	}
	return s % 20
}

type ConnectionManager struct {
	buckets      []*Bucket
	dispatchChan chan PushJob // 待分发消息队列
	Epoller      []*Epoller
}

func GetConnectionInfo() int {
	var count int
	for _, v := range CM.buckets {
		count += v.Count()
	}
	return count
}
func NewConnectionManager() {
	CM = &ConnectionManager{
		buckets:      make([]*Bucket, global.Config.Bucket.BucketCount),
		dispatchChan: make(chan PushJob, global.Config.Bucket.DispatchChanSize),
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

//将连接移除事件监听中
func (c *ConnectionManager) removeEpollerConn(id int64) error {
	i := id % 4
	return c.Epoller[i].Remove(int(id))
}

//将连接加入到事件监听中
func (c *ConnectionManager) addEpollerConn(id int64) error {
	i := id % 4
	return c.Epoller[i].Add(int(id))
}

//分发数据到到所有bucket的jobchan中
func (c *ConnectionManager) dispatchToBucket() {

	for data := range c.dispatchChan {
		//不同的room并发,room中的数据是有序的
		rid := data.GetRoomID()
		for _, v := range c.buckets {
			v.messageChan[rid] <- data
		}
	}
}

// PushRoom 推送到room
func (c *ConnectionManager) PushRoom(room string, data []byte) {
	job := PushJob{
		PushType: PushRoom,
		roomID:   room,
		data:     data,
	}

	c.dispatchChan <- job
}

// PushAll 推送给所有
func (c *ConnectionManager) PushAll(data []byte) {
	job := PushJob{
		PushType: PushAll,
		data:     data,
	}
	c.dispatchChan <- job
}

// PushPerson 推送给所有
func (c *ConnectionManager) PushPerson(uid string, data []byte) {
	job := PushJob{
		PushType: PushPerson,
		uid:      uid,
		data:     data,
	}
	c.dispatchChan <- job
}

// AddConnection 添加连接
func (c *ConnectionManager) AddConnection(conn *Connection) {
	b := c.getBucket(conn.ID)
	b.AddConn(conn)
}

// NotifyRead 通知连接读取
func (c *ConnectionManager) NotifyRead(fd int) {
	b := c.getBucket(int64(fd))
	i := fd % 10
	b.notify[i] <- fd
}

func (c *ConnectionManager) CloseConnection(fd int) {
	b := c.getBucket(int64(fd))
	conn, ok := b.GetConnection(int64(fd))
	if ok {
		conn.Close()
	}
}

// DelConnection 删除连接
func (c *ConnectionManager) DelConnection(conn *Connection) {
	b := c.getBucket(conn.ID)
	b.DelConn(conn)
}

func (c *ConnectionManager) JoinRoom(roomID string, conn *Connection) {
	if ok := conn.isSubbed(roomID); ok {
		return
	}
	//新增连接上的room
	conn.subRoom(roomID)
	b := c.getBucket(conn.ID)
	b.JoinRoom(roomID, conn)
}
func (c *ConnectionManager) LeaveRoom(roomID string, conn *Connection) {
	//删除连接上的room
	conn.unSubRoom(roomID)
	b := c.getBucket(conn.ID)
	b.LeaveRoom(roomID, conn)
}
func (c *ConnectionManager) LoginPrivateRoom(roomID string, conn *Connection) {
	//删除连接上的room
	ok := conn.isSubbed(roomID)
	if ok {
		return
	}
	b := c.getBucket(conn.ID)
	b.LeaveRoom(roomID, conn)
}

func (c *ConnectionManager) getBucket(connID int64) *Bucket {
	return c.buckets[connID%int64(len(c.buckets))]
}
