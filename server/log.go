package server

import (
	"fmt"
	"log"
)

func LogInfo(msg string, args ...interface{}) {
	Log("INFO", msg, args...)
}

func LogError(msg string, args ...interface{}) {
	Log("ERROR", msg, args...)
}

func LogCritical(msg string, args ...interface{}) {
	Log("CRIT", msg, args...)
	panic(fmt.Sprintf(msg, args...))
}

func Log(level string, msg string, args ...interface{}) {
	log.Printf("%v %v\n", level, fmt.Sprintf(msg, args...))
}
