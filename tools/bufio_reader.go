package tools

import (
	"errors"
	"github.com/spf13/cast"
	"io"
)

const (
	defaultBufSize = 4096
)

var (
	ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
	ErrInvalidUnreadRune = errors.New("bufio: invalid use of UnreadRune")
	ErrBufferFull        = errors.New("bufio: buffer full")
	ErrNegativeCount     = errors.New("bufio: negative count")
)

// Buffered input.

// Reader implements buffering for an io.Reader object.
type Reader struct {
	buf            []byte
	rd             io.Reader // reader provided by the client
	r, w           int       // buf read and write positions
	err            error
	lastByte       int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize   int // size of last rune read for UnreadRune; -1 means invalid
	lastMessagePos int //上一条消息读取的位置。
}

const minReadBufferSize = 16
const maxConsecutiveEmptyReads = 100

// NewReaderSize returns a new Reader whose buffer has at least the specified
// size. If the argument io.Reader is already a Reader with large enough
// size, it returns the underlying Reader.
func NewReaderSize(rd io.Reader, size int) *Reader {
	// Is it already a Reader?
	b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

// NewReader returns a new Reader whose buffer has the default size.
func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

// Size returns the size of the underlying buffer in bytes.
func (b *Reader) Size() int { return len(b.buf) }

// Reset discards any buffered data, resets all state, and switches
// the buffered reader to read from r.
// Calling Reset on the zero value of Reader initializes the internal buffer
// to the default size.
func (b *Reader) Reset(r io.Reader) {
	if b.buf == nil {
		b.buf = make([]byte, defaultBufSize)
	}
	b.reset(b.buf, r)
}

func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf:          buf,
		rd:           r,
		lastByte:     -1,
		lastRuneSize: -1,
	}
}

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

// fill reads a new chunk into the buffer.
func (b *Reader) fill() {
	// Slide existing data to beginning.
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}

	if b.w >= len(b.buf) {
		panic("bufio: tried to fill full buffer")
	}

	// Read new data: try a limited number of times.
	for i := maxConsecutiveEmptyReads; i > 0; i-- {
		n, err := b.rd.Read(b.buf[b.w:])
		if n < 0 {
			panic(errNegativeRead)
		}
		b.w += n
		if err != nil {
			b.err = err
			return
		}
		if n > 0 {
			return
		}
	}
	b.err = io.ErrNoProgress
}

// UpdateLastMessagePos 更新读完上一条数据的位置。
func (b *Reader) UpdateLastMessagePos() {
	b.lastMessagePos = b.r
}

// Rewind  回退到读完上一条数据的位置。
func (b *Reader) Rewind() {
	b.r = b.lastMessagePos
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

// Read reads data into p.
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying Reader,
// hence n may be less than len(p).
// To read exactly len(p) bytes, use io.ReadFull(b, p).
// At EOF, the count will be zero and err will be io.EOF.
func (b *Reader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		if b.Buffered() > 0 {
			return 0, nil
		}
		return 0, b.readErr()
	}
	if b.r == b.w {
		if b.err != nil {
			return 0, b.readErr()
		}

		//如果要读取数据，则要调整缓存数据的位置，从零开始。
		if b.lastMessagePos > 0 {
			n := copy(b.buf, b.buf[b.lastMessagePos:b.w])
			b.r = n
			b.w = n
			b.lastMessagePos = 0
		}

		//缓冲区无法存储完整的消息，执行扩容
		if b.w == b.Size() {
			b.grow()
		}

		n, b.err = b.rd.Read(b.buf[b.w:])
		if n < 0 {
			panic(errNegativeRead)
		}
		if n == 0 {
			return 0, b.readErr()
		}
		b.w += n
	}

	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = -1
	return n, nil
}

func (b *Reader) grow() {

	size := float64(len(b.buf)) * 1.5
	if b.Size() > 10240 {
		size = float64(len(b.buf)) * 1.25
	}

	buf := make([]byte, cast.ToInt64(size))
	n := copy(buf, b.buf)
	b.buf = buf
	b.w = n
}

// Buffered returns the number of bytes that can be read from the current buffer.
func (b *Reader) Buffered() int { return b.w - b.r }
