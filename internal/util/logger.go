package util

import (
	"log"
	"os"
)

const PathToLogs = "../../logs/app.log"

var (
	logFile *os.File
	logChan = make(chan string, 1000)
)

func StartLogWriter(){
	var err error
	logFile, err = os.OpenFile(PathToLogs, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("Can't load logger: %v", err)
	}

	setLogSetting()

	defer logFile.Close()

	for msg := range logChan {
		log.Print(msg)
	}
}

func setLogSetting(){

	log.SetOutput(logFile)

	log.SetPrefix("LOG: ")

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func CloseLogChan(){
	close(logChan)
}

func LogWrite(str string){
	logChan <- str
}