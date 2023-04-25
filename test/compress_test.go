package test

import (
	"compress/flate"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/luxun9527/gpush/internal/socket/model"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
)

func TestCompressServer(t *testing.T) {
	http.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		var httpUpgrade ws.HTTPUpgrader
		e := wsflate.Extension{
			Parameters: wsflate.DefaultParameters,
		}
		httpUpgrade.Negotiate = e.Negotiate
		conn, _, _, err := httpUpgrade.Upgrade(request, writer)
		if err != nil {
			log.Println(err)
			return
		}
		go func() {
			defer conn.Close()

			for {
				frame, err := ws.ReadFrame(conn)
				frame = ws.UnmaskFrameInPlace(frame)

				if err != nil {
					log.Println(err)
					return
				}
				compressed, err := wsflate.IsCompressed(frame.Header)
				if err != nil {
					log.Println(err)
				}
				if compressed {
					// Note that even after successful negotiation of
					// compression extension, both sides are able to send
					// non-compressed messages.
					frame, err = wsflate.DecompressFrame(frame)
					if err != nil {
						log.Println(err)
						return
					}
					log.Println(frame.Header.OpCode)
					log.Println(string(frame.Payload))
					message := model.NewMessage(ws.OpText, frame.Payload)
					data, _ := message.ToCompressBytes()
					conn.Write(data)
				}
				//log.Println(string(d1),op)

			}
		}()

	})

	http.ListenAndServe(":34442", nil)
}
func TestCompressTo(t *testing.T) {
	var DefaultHelper = wsflate.Helper{
		Compressor: func(w io.Writer) wsflate.Compressor {
			// No error can be returned here as NewWriter() doc says.
			f, _ := flate.NewWriter(w, 9)
			return f
		},
		Decompressor: func(r io.Reader) wsflate.Decompressor {
			return flate.NewReader(r)
		},
	}

	data := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	}
	compress, err := DefaultHelper.Compress(data)
	if err != nil {
		log.Println(err)
	}
	//	0, 3, 0, 252, 255, 97, 98, 99, 0,
	// [0 3 0 252 255 97 98 99 0 0 0 255 255 1]
	log.Println((compress))
}

func TestCompress(t *testing.T) {
	data := []byte{
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
	}
	frame, err := model.NewMessage(ws.OpText, data).ToCompressBytes()
	if err != nil {
		log.Println("err", err)
		return
	}
	log.Println(frame)
	log.Println(len(frame))
}

func TestFlateServer(t *testing.T) {

	http.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		var httpUpgrade ws.HTTPUpgrader
		e := wsflate.Extension{
			Parameters: wsflate.DefaultParameters,
		}
		httpUpgrade.Negotiate = e.Negotiate
		conn, _, _, err := httpUpgrade.Upgrade(request, writer)
		if err != nil {
			log.Println(err)
			return
		}
		go func() {
			defer conn.Close()

			for {
				frame, err := ws.ReadFrame(conn)
				frame = ws.UnmaskFrameInPlace(frame)

				if err != nil {
					log.Println(err)
					return
				}
				compressed, err := wsflate.IsCompressed(frame.Header)
				if err != nil {
					log.Println(err)
				}
				if compressed {
					// Note that even after successful negotiation of
					// compression extension, both sides are able to send
					// non-compressed messages.
					frame, err = wsflate.DecompressFrame(frame)
					if err != nil {
						log.Println(err)
						return
					}
					log.Println(frame.Header.OpCode)
					log.Println(string(frame.Payload))
					message := model.NewMessage(ws.OpText, frame.Payload)
					data, _ := message.ToBytes()
					conn.Write(data)
				}

			}
		}()

	})

	http.ListenAndServe(":34442", nil)
}

func TestFlateServer1(t *testing.T) {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	e := wsflate.Extension{
		// We are using default parameters here since we use
		// wsflate.{Compress,Decompress}Frame helpers below in the code.
		// This assumes that we use standard compress/flate package as flate
		// implementation.
		Parameters: wsflate.DefaultParameters,
	}
	u := ws.Upgrader{
		Negotiate: e.Negotiate,
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		// Reset extension after previous upgrades.
		e.Reset()

		_, err = u.Upgrade(conn)
		if err != nil {
			log.Printf("upgrade error: %s", err)
			continue
		}
		if _, ok := e.Accepted(); !ok {
			log.Printf("didn't negotiate compression for %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		go func() {
			defer conn.Close()
			for {
				frame, err := ws.ReadFrame(conn)
				if err != nil {
					// Handle error.
					return
				}

				frame = ws.UnmaskFrameInPlace(frame)
				compressed, err := wsflate.IsCompressed(frame.Header)

				if compressed {
					// Note that even after successful negotiation of
					// compression extension, both sides are able to send
					// non-compressed messages.
					frame, err = wsflate.DecompressFrame(frame)
					if err != nil {
						// Handle error.
						return
					}
				}

				// Do something with frame...

				ack := ws.NewTextFrame([]byte{'a', 'b', 'c'})

				// Compress response unconditionally.
				ack, err = wsflate.CompressFrame(ack)
				if err != nil {
					// Handle error.
					return
				}
				if err = ws.WriteFrame(conn, ack); err != nil {
					// Handle error.
					return
				}
			}

		}()
	}
}
