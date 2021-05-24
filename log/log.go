package log

import (
	"fmt"
	"log"
	"os"
)

const flags = log.Ldate | log.Ltime | log.Lshortfile

var s *log.Logger

func init() {
	s = log.New(os.Stderr, "", flags)
}

func Output(calldepth int, v ...interface{}) {
	s.Output(calldepth+2, fmt.Sprintln(v...))
}

func Debug(v ...interface{}) {
	Output(1, v...)
}

func Info(v ...interface{}) {
	Output(1, v...)
}

func Error(v ...interface{}) {
	Output(1, v...)
}
