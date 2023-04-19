package config

type Connection struct {
	WriteRate  int  `mapstructure:"writeRate"`
	TimeOut    int  `mapstructure:"timeOut"`
	IsCompress bool `mapstructure:"isCompress"`
	WriteBuf   int  `mapstructure:"writeBuf"`
	ReadBuf    int  `mapstructure:"readBuf"`
}
