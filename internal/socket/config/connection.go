package config

type Connection struct {
	WriteRate         int
	TimeOut           int
	IsCompress        bool
	WriteBuf          int
	ReadBuf           int
	EnableWriteBuffer bool
	EableEpoller      bool
}
