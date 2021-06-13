package exasol

import (
	"errors"
	"log"
	"os"
)

var errorLogger = Logger(log.New(os.Stderr, "[exasol] ", log.LstdFlags|log.Lshortfile))

// Logger is used to log critical error messages.
type Logger interface {
	Print(v ...interface{})
}

// SetLogger is used to set the logger for critical errors.
// The initial logger is os.Stderr.
func SetLogger(logger Logger) error {
	if logger == nil {
		return errors.New("logger is nil")
	}
	errorLogger = logger
	return nil
}
