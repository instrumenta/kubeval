package log

import (
	"fmt"

	"github.com/fatih/color"
)

func Info(message ...interface{}) {
	fmt.Println(message...)
}

func Success(message ...interface{}) {
	green := color.New(color.FgGreen)
	green.Println(message...)
}

func Warn(message ...interface{}) {
	yellow := color.New(color.FgYellow)
	yellow.Println(message...)
}

func Error(message ...interface{}) {
	red := color.New(color.FgRed)
	red.Println(message...)
}
