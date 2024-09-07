package log

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
	"time"
)

type logger struct {
	writer io.Writer
	output *log.Logger
	info   *log.Logger
	warn   *log.Logger
	error  *log.Logger
}

// 默认日志级别
var logLevel = InfoLevel

const (
	InfoLevel int = iota
	WarnLevel
	ErrorLevel
)

var loggerInstance *logger
var writers []io.Writer
var once sync.Once

func AddLogWriter(w io.Writer) {
	writers = append(writers, w)
}

func GetLogWriter() io.Writer {
	return getLogger().writer
}

func GetLogWriterExceptOsStdout() io.Writer {
	if len(writers) == 0 {
		return nil
	}
	var newWriters []io.Writer
	for _, writer := range writers {
		if writer == os.Stdout {
			continue
		}
		newWriters = append(newWriters, writer)
	}
	if len(newWriters) == 0 {
		return nil
	}
	return io.MultiWriter(newWriters...)
}

func SetLogLevel(level int) {
	logLevel = level
}

func initDefaultWriter() {
	logfile := fmt.Sprintf("./davinci_%s.log", time.Now().Format("2006-01-02"))
	if fileWriter, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600); err != nil {
		log.Println(fmt.Sprintf("create log file error: %v", err))
	} else {
		writers = append(writers, fileWriter)
	}
	writers = append(writers, os.Stdout)
}

func getLogger() *logger {
	once.Do(func() {
		if writers == nil {
			initDefaultWriter()
		}
		logWriter := io.MultiWriter(writers...)
		loggerInstance = &logger{
			writer: logWriter,
			output: log.New(logWriter, "", 0),
			info:   log.New(logWriter, "[info] ", log.Lmsgprefix|log.Ldate|log.Ltime),
			warn:   log.New(logWriter, "[warn] ", log.Lmsgprefix|log.Ldate|log.Ltime),
			error:  log.New(logWriter, "[error] ", log.Lmsgprefix|log.Ldate|log.Ltime),
		}
	})
	return loggerInstance
}

func Close() {
	for _, writer := range writers {
		if writer != nil {
			if reflect.TypeOf(writer) == reflect.TypeOf(&os.File{}) {
				writer.(*os.File).Close()
			}
		}
	}
}

/**
* 按照数据库SQL Table的格式输出数据
 */
func Output(v interface{}) {
	l := getLogger()
	if v == nil {
		return
	}
	if data, ok := v.([]string); ok {
		logArray(data, l.output.Writer())
	} else if data, ok := v.([][]string); ok {
		logTable(data, l.output.Writer())
	} else {
		l.output.Println(v)
	}

}

func Info(v ...any) {
	if logLevel <= InfoLevel {
		getLogger().info.Println(v...)
	}

}

func Warn(v ...any) {
	if logLevel <= WarnLevel {
		getLogger().warn.Println(v...)
	}
}

func Error(v ...any) {
	if logLevel <= ErrorLevel {
		getLogger().error.Println(v...)
	}
}

func logTable(result [][]string, writer io.Writer) {
	if result == nil || len(result) == 0 {
		return
	}
	table := tablewriter.NewWriter(writer)
	table.SetHeader(result[0])

	for i := 1; i < len(result); i++ {
		table.Append(result[i])
	}
	table.Render()
}

func logArray(result []string, writer io.Writer) {
	if result == nil || len(result) == 0 {
		return
	}
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{result[0]})
	for i := 1; i < len(result); i++ {
		table.Append([]string{result[i]})
	}
	table.Render()
}
