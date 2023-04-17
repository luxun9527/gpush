package config

type Config struct {
	Bucket     Bucket       `mapstructure:"bucket"`
	Server     Server       `mapstructure:"server"`
	Logger     LoggerConfig `mapstructure:"logger"`
	Proxy      Proxy        `mapstructure:"proxy"`
	Connection Connection   `mapstructure:"connection"`
}
