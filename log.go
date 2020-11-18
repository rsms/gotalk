package gotalk

import "log"

// LoggerFunc is the signature of log functions.
// s is a related socket, or nil if none apply to the error.
type LoggerFunc func(s *Sock, format string, args ...interface{})

var (
	// ErrorLogger is called when an error occurs during writing or reading (rare)
	ErrorLogger LoggerFunc = DefaultLoggerFunc

	// HandlerErrorLogger is called when a handler function either panics or returns an error
	HandlerErrorLogger LoggerFunc = DefaultLoggerFunc
)

// DefaultLoggerFunc forwards the message to Go's "log" package; log.Printf(format, args...)
func DefaultLoggerFunc(s *Sock, format string, args ...interface{}) {
	log.Printf(format, args...)
}
