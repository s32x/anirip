package common

import "github.com/fatih/color"

// Logger is a struct used for logging
type Logger struct{ prefix string }

// NewLogger declares a new Logger pointer
func NewLogger() *Logger { return &Logger{prefix: "[anirip] "} }

// Start logs the start of a procedure
func (l *Logger) Cyan(format string, a ...interface{}) {
	color.Cyan(l.prefix+format, a...)
}

// Info logs basic information
func (l *Logger) Info(format string, a ...interface{}) {
	color.White(l.prefix+format, a...)
}

// Warn logs warnings
func (l *Logger) Warn(format string, a ...interface{}) {
	color.Yellow(l.prefix+format, a...)
}

// Success logs a success message
func (l *Logger) Success(format string, a ...interface{}) {
	color.Green(l.prefix+format, a...)
}

// Error logs errors
func (l *Logger) Error(err error) {
	color.Red(l.prefix + err.Error())
}
