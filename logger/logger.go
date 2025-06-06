package logger

import (
	"log"
	"os"
)

var InfoLog *log.Logger

func InitLogger(filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	InfoLog = log.New(file, "", log.LstdFlags|log.Lmsgprefix)
	InfoLog.SetPrefix("INFO: ")
}
