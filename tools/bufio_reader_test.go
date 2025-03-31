package tools

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestBuf(t *testing.T) {
	r := NewReaderSize(strings.NewReader("1234567890123456789012345678901234567890"), 16)

	buf := make([]byte, 25)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			r.Rewind()
		}
	}
	t.Logf("lastMessagePos=%v r=%v w=%v data=%v ", r.lastMessagePos, r.r, r.w, string(buf[:n]))
	//lastMessagePos=0 r=25 w=36 data=1234567890123456789012345
	r.UpdateLastMessagePos()
	n, err = io.ReadFull(r, buf)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			r.Rewind()
			t.Logf("rewind lastMessagePos=%v r=%v w=%v size=%v,bufData=%v", r.lastMessagePos, r.r, r.w, r.Size(), string(r.buf[r.r:r.w]))
			//rewind lastMessagePos=0 r=0 w=15 size=36,bufData=678901234567890
			return
		}
	}
	t.Logf("lastMessagePos=%v r=%v w=%v  size=%v data=%v", r.lastMessagePos, r.r, r.w, r.Size(), string(buf[:n]))
}
