package initialize

import (
	"github.com/luxun9527/gpush/internal/proxy/global"
	"github.com/luxun9527/zlog"
	"github.com/spf13/viper"
	"log"
)

func InitConfig(addr string) {
	viperConfig := viper.New()
	viperConfig.SetConfigFile(addr)
	if err := viperConfig.ReadInConfig(); err != nil {
		log.Panicf("初始化日志失败 err = %s", err.Error())
	}
	if err := viperConfig.Unmarshal(&global.Config, viper.DecodeHook(zlog.StringToLogLevelHookFunc())); err != nil {
		log.Panicf("unmarshalKey ws failed err = %s", err.Error())
	}
	zlog.InitDefaultLogger(&global.Config.Logger)

}
