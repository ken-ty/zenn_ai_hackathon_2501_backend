package logging

import (
	"log"
	"os"
	"strings"
)

// LogLevel はログレベルを表す型
type LogLevel int

const (
	// DEBUG は最も詳細なログレベル
	DEBUG LogLevel = iota
	// INFO は一般的な情報のログレベル
	INFO
	// WARN は警告レベル
	WARN
	// ERROR はエラーレベル
	ERROR
)

var (
	// currentLevel は現在のログレベル
	currentLevel = INFO
)

// init は環境変数からログレベルを設定します
func init() {
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(level) {
	case "DEBUG":
		currentLevel = DEBUG
	case "INFO":
		currentLevel = INFO
	case "WARN":
		currentLevel = WARN
	case "ERROR":
		currentLevel = ERROR
	}

	// テスト実行時は自動的にWARNレベルに設定
	if strings.HasSuffix(os.Args[0], ".test") {
		currentLevel = WARN
	}
}

// Debug はDEBUGレベルのログを出力します
func Debug(format string, v ...interface{}) {
	if currentLevel <= DEBUG {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info はINFOレベルのログを出力します
func Info(format string, v ...interface{}) {
	if currentLevel <= INFO {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn はWARNレベルのログを出力します
func Warn(format string, v ...interface{}) {
	if currentLevel <= WARN {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error はERRORレベルのログを出力します
func Error(format string, v ...interface{}) {
	if currentLevel <= ERROR {
		log.Printf("[ERROR] "+format, v...)
	}
}

// SetLevel は明示的にログレベルを設定します
func SetLevel(level LogLevel) {
	currentLevel = level
}
