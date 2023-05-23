package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// 这个颜色的标记符在 python 中也通用
var (

	// Lshortfile 表示显示文件名及行号; Llongfile 表示显示完整路径及文件名及行号；
	errorLog = log.New(os.Stdout, "\033[31m[ERROR]\033[0m", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[INFO ]\033[0m", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

// 重新定义 log 的一些打印方法，或者说取别名
var (
	Error = errorLog.Println
	Info  = infoLog.Println
	// Errorf Infof 以 format 格式打印
	Errorf = errorLog.Printf
	Infof  = infoLog.Printf
)

// 支持的日志层级 InfoLevel, ErrorLevel, Disabled
// 三个层级声明为三个常量，通过控制 Output，来控制日志是否打印
// iota 是 go 特殊的可变常量
const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

// SetLevel 设置日志等级控制
func SetLevel(level int) {
	// 上锁
	mu.Lock()
	// 执行完毕后解锁
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	//如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志
	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}

	//如果设置为 InfoLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志
	if InfoLevel < level {
		infoLog.SetOutput(ioutil.Discard)
	}
}
