package util

import (
	"fmt"
	"os"
)

var IsTraceEnabled bool

func Write(format string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, format, msg...)
}

func Writeln(format string, msg ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, msg...))
}

func Traceln(format string, msg ...interface{}) {
	if IsTraceEnabled {
		fmt.Fprintln(os.Stderr, fmt.Sprintf(format, msg...))
	}
}

func Exit(err error) {
	if err != nil {
		Writeln(err.Error())
	}
	os.Exit(1)
}
