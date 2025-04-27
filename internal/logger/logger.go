package logger

import (
	"fmt"
	"os"
	"time"
	"github.com/fatih/color"
)

var debugEnabled bool
var component string

// Init sets the debug flag and component tag for logging
func Init(debug bool, comp string) {
	debugEnabled = debug
	component = comp
}

// logMessage formats and prints a timestamped, colored log line
func logMessage(level string, attr color.Attribute, format string, v ...interface{}) {
	ts := time.Now().Format(time.RFC3339)
	fmt.Print(ts, " ", component, " ")
	color.New(attr).Print(level)
	fmt.Print(" ")
	fmt.Printf(format+"\n", v...)
}

// Info logs informational messages (green)
func Info(format string, v ...interface{}) {
	logMessage("INFO", color.FgGreen, format, v...)
}

// Debug logs debug messages only when enabled (yellow)
func Debug(format string, v ...interface{}) {
	if debugEnabled {
		logMessage("DEBUG", color.FgYellow, format, v...)
	}
}

// Error logs error messages (red)
func Error(format string, v ...interface{}) {
	logMessage("ERROR", color.FgRed, format, v...)
}

// Fatalf logs a formatted fatal error and exits (red)
func Fatalf(format string, v ...interface{}) {
	logMessage("FATAL", color.FgRed, format, v...)
	os.Exit(1)
}

// Fatal logs a fatal error and exits (red)
func Fatal(v ...interface{}) {
	// join args into one message
	msg := fmt.Sprint(v...)
	logMessage("FATAL", color.FgRed, "%s", msg)
	os.Exit(1)
}

// Printf logs a formatted message at INFO level
func Printf(format string, v ...interface{}) {
	logMessage("INFO", color.FgWhite, format, v...)
}

// Println logs a line message at INFO level
func Println(v ...interface{}) {
	msg := fmt.Sprint(v...)
	logMessage("INFO", color.FgWhite, "%s", msg)
}
