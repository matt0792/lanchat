package logger

import (
	"log"
	"os"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone
)

var (
	currentLevel = LevelInfo
	mu           sync.RWMutex
)

func SetLevel(level Level) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

func Debug(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= LevelDebug {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func Info(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= LevelInfo {
		log.Printf("[INFO] "+format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= LevelWarn {
		log.Printf("[WARN] "+format, args...)
	}
}

func Error(format string, args ...interface{}) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= LevelError {
		log.Printf("[ERROR] "+format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] "+format, args...)
}

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)
}
