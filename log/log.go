package log

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	multierror "github.com/hashicorp/go-multierror"
)

func Success(message ...string) {
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s - %v\n", green("PASS"), strings.Join(message, " "))
}

func Warn(message ...string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("%s - %v\n", yellow("WARN"), strings.Join(message, " "))
}

func Error(message error) {
	if merr, ok := message.(*multierror.Error); ok {
		for _, serr := range merr.Errors {
			Error(serr)
		}
	} else {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s - %v\n", red("ERR "), message)
	}
}
