package log /* import "s32x.com/anirip/common/log" */

import "github.com/fatih/color"

const prefix = "[anirip] "

// Cyan logs important information in Cyan
func Cyan(format string, a ...interface{}) {
	color.Cyan(prefix+format, a...)
}

// Info logs basic information
func Info(format string, a ...interface{}) {
	color.White(prefix+format, a...)
}

// Warn logs warnings
func Warn(format string, a ...interface{}) {
	color.Yellow(prefix+format, a...)
}

// Success logs a success message
func Success(format string, a ...interface{}) {
	color.Green(prefix+format, a...)
}

// Error logs errors
func Error(err error) {
	color.Red(prefix + err.Error())
}
