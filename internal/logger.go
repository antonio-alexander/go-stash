package internal

import "fmt"

type Logger interface {
	Printf(format string, a ...any) (n int, err error)
}

type logger struct{}

func NewLogger() Logger {
	return &logger{}
}

func (l *logger) Printf(format string, a ...any) (int, error) {
	return fmt.Printf(format, a...)
}
