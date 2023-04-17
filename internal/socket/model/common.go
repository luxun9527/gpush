package model

import (
	"bytes"
	gws "github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
)

type Message struct {
	messageType gws.OpCode
	data        []byte
}

func NewMessage(code gws.OpCode, data []byte) Message {
	return Message{
		messageType: code,
		data:        data,
	}
}

func (m Message) ToBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 100))
	frame := gws.NewFrame(m.messageType, true, m.data)
	if err := gws.WriteFrame(buf, frame); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m Message) ToCompressBytes() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 100))
	frame := gws.NewFrame(m.messageType, true, m.data)
	compressFrame, err := wsflate.CompressFrame(frame)
	if err != nil {
		return nil, err
	}
	if err := gws.WriteFrame(buf, compressFrame); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
