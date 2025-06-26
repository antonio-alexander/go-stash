package internal

import (
	"fmt"

	"github.com/antonio-alexander/go-stash"
)

type logger struct{}

func NewLogger() stash.Logger {
	return &logger{}
}

func (l *logger) Printf(format string, a ...any) {
	fmt.Printf(format, a...)
}
