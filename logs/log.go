package loger

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

const (
	LogSuccess   = "日志引擎初始化成功"
	ErrLogStart  = "日志引擎初始化失败"
	ErrLogConfig = "日志引擎初始化失败(日志配置文件转换失败)："
)

// InitLog 初始化日志引擎
func InitLog() {
	logTime := fmt.Sprintf("log_%s.log", time.Now().Format("20060102"))
	var filePath string
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		filePath = filepath.Join("./", "logs", logTime)
	} else {
		filePath = filepath.Join(".\\", "logs", logTime)
	}
	var logConfig = map[string]interface{}{ // 配置日志引擎
		"filename": filePath,          // 保存的文件名
		"maxLines": 100000,            // 每个文件保存的最大行数
		"maxSize":  256 * 1024 * 1024, // 每个文件保存的最大尺寸
		"daily":    true,              // 是否按照每天分割日志
		"maxDays":  7,                 // 文件最多保存多少天
		"rotate":   true,              // 是否开启日志分割
		"level":    7,                 // 日志保存的时候的级别,日志保存级别：1-Alert,2-Critical,3-Error,4-Warning,5-Notice,6-Info,7-Debug;(默认值：7)
	}
	logConfStr, err := json.Marshal(logConfig) // 转换配置为字符串
	if err != nil {
		fmt.Println(ErrLogConfig, err.Error()) //日志配置文件转换失败
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(logConfStr)) // 开启文件日志记录
	if err != nil {
		logs.Error(ErrLogStart, err.Error()) //初始化日志失败
		return
	}
	logs.EnableFuncCallDepth(true) // 开启日志行号及文件打印
	logs.Async()                   // 开启异步日志记录
	logs.Info(LogSuccess)          // 初始化日志初始化成功
}
