package logger

import (
	"log"
	"os"
)

var IsDebug = os.Getenv("DEBUG_MODE") == "true"

func Info(msg string, args ...interface{}) {
	log.Printf("ℹ️ "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	log.Printf("❌ "+msg, args...)
}

func Debug(msg string, args ...interface{}) {
	if IsDebug {
		log.Printf("🐞 "+msg, args...)
	}
}
