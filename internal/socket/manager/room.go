package manager

import (
	"sync"
)

type Room struct {
	roomID      string
	id2Conn     map[int64]*Connection
	id2ConnLock sync.RWMutex
}

func NewRoom(roomID string) *Room {
	return &Room{
		roomID:      roomID,
		id2Conn:     make(map[int64]*Connection, 200),
		id2ConnLock: sync.RWMutex{},
	}
}

func (r *Room) Join(conn *Connection) {
	r.id2ConnLock.Lock()
	r.id2Conn[conn.ID] = conn
	r.id2ConnLock.Unlock()
}

func (r *Room) Leave(conn *Connection) {
	r.id2ConnLock.Lock()
	delete(r.id2Conn, conn.ID)
	r.id2ConnLock.Unlock()
}
func (r *Room) Count() int {
	r.id2ConnLock.RLock()
	defer r.id2ConnLock.RUnlock()
	return len(r.id2Conn)
}

func (r *Room) Push(data []byte) {
	r.id2ConnLock.RLock()
	defer r.id2ConnLock.RUnlock()
	for _, conn := range r.id2Conn {
		conn.Send(data)
	}
}
