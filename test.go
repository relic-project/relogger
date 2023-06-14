package main

import (
	"github.com/relic-project/relogger/writers"
	"time"
)

func main() {
	writer := writers.DefaultFileLoggingWriter
	writer.Write([]byte("hello world"))
	time.Sleep(5 * time.Second)
}
