package tools

import (
	"encoding/binary"
	"errors"
	"github.com/gobwas/ws"
	"io"
)

const (
	bit0 = 0x80
	bit1 = 0x40
	bit2 = 0x20
	bit3 = 0x10
	bit4 = 0x08
	bit5 = 0x04
	bit6 = 0x02
	bit7 = 0x01

	len7  = int64(125)
	len16 = int64(^(uint16(0)))
	len64 = int64(^(uint64(0)) >> 1)
)

var (
	MessageTooLarge = errors.New("message size over max limit")
)

type WebSocketReader struct {
	bufioReader    *Reader
	pos            int
	maxMessageSize int64
}

func NewWebSocketReader(rd io.Reader, size int) *WebSocketReader {
	return &WebSocketReader{
		bufioReader:    NewReaderSize(rd, size),
		pos:            0,
		maxMessageSize: int64(size),
	}
}

// ReadHeader reads a frame header from r.
func (wr *WebSocketReader) ReadHeader(r io.Reader) (h ws.Header, err error) {
	// Make slice of bytes with capacity 12 that could hold any header.
	//
	// The maximum header size is 14, but due to the 2 hop reads,
	// after first hop that reads first 2 constant bytes, we could reuse 2 bytes.
	// So 14 - 2 = 12.
	bts := make([]byte, 2, ws.MaxHeaderSize-2)
	var n int
	// Prepare to hold first 2 bytes to choose size of next read.
	n, err = io.ReadFull(r, bts)
	if err != nil {
		wr.pos += n
		return
	}
	wr.pos += n
	h.Fin = bts[0]&bit0 != 0
	h.Rsv = (bts[0] & 0x70) >> 4
	h.OpCode = ws.OpCode(bts[0] & 0x0f)

	var extra int

	if bts[1]&bit0 != 0 {
		h.Masked = true
		extra += 4
	}

	length := bts[1] & 0x7f
	switch {
	case length < 126:
		h.Length = int64(length)

	case length == 126:
		extra += 2

	case length == 127:
		extra += 8

	default:
		err = ws.ErrHeaderLengthUnexpected
		return
	}

	if extra == 0 {
		return
	}

	// Increase len of bts to extra bytes need to read.
	// Overwrite first 2 bytes that was read before.
	bts = bts[:extra]
	n, err = io.ReadFull(r, bts)
	if err != nil {
		wr.pos += n
		return
	}
	wr.pos += n
	switch {
	case length == 126:
		h.Length = int64(binary.BigEndian.Uint16(bts[:2]))
		bts = bts[2:]

	case length == 127:
		if bts[0]&0x80 != 0 {
			err = ws.ErrHeaderLengthMSB
			return
		}
		h.Length = int64(binary.BigEndian.Uint64(bts[:8]))
		bts = bts[8:]
	}

	if h.Masked {
		copy(h.Mask[:], bts)
	}
	return
}

func (wr *WebSocketReader) ReadFrame() (f ws.Frame, err error) {
	f.Header, err = wr.ReadHeader(wr.bufioReader)
	if err != nil {
		return
	}
	if f.Header.Length > wr.maxMessageSize {
		err = MessageTooLarge
		return
	}
	if f.Header.Length > 0 {
		// int(f.Header.Length) is safe here cause we have
		// checked it for overflow above in ReadHeader.
		f.Payload = make([]byte, int(f.Header.Length))
		var n int
		n, err = io.ReadFull(wr.bufioReader, f.Payload)
		if err != nil {
			wr.pos += n
		} else {
			//沒出现错误也要将剩下的数据调整位置，从零开始
			wr.bufioReader.adjustPos()
			wr.pos = 0
		}
	}

	return
}

func (wr WebSocketReader) Reset() error {
	return wr.bufioReader.GoBackN(wr.pos)
}
