package request

type Code uint8

const (
	SubPublic Code = iota + 1
	UnsubPublic
	Login
	Logout
	SubPrivate
	UnSubSubPrivate
)

type Message struct {
	Code  Code        `json:"code"`
	Topic string      `json:"topic"`
	Data  interface{} `json:"data"`
}
