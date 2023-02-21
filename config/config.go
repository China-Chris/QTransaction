package c

import (
	"context"
	"github.com/spf13/viper"
	"sync"
)

var (
	config  GlobalConfig
	rConfig sync.RWMutex
)

// MysqlConfig mysql配置参数
type MysqlConfig struct {
	User     string
	Password string
	Ip       string
	Port     string
	DbName   string
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	Port    string
	Mysql   MysqlConfig
	Redis   Redis
	Message Message
	Debug   bool
	Proxy   string
}

// Config 返回配置文件
func Config() GlobalConfig {
	rConfig.RLock()
	configCopy := config
	rConfig.RUnlock()
	return configCopy
}

type Redis struct {
	Host   string // 数据库连接地址
	Port   int64  // 数据库连接端口
	DbName int    // 数据库名称
	Passwd string // 数据库密码
}

type Message struct {
	Message     string // 短信发送地址
	MessagePath string // 短信path
}

var Ctx = context.Background()

//ParseConfig 加载配置文件
func ParseConfig(cfg string) {
	viper.SetConfigFile(cfg)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}
}
