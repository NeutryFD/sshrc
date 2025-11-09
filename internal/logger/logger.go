package logger

import (
	"fmt"
	"time"
)

// LogStep prints a timestamped log message
func LogStep(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, message)
}
