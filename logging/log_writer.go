package logging

import (
	"log"
	"strings"
)

// Logger logs messages
type Logger interface {
	Log(msg string)
}

// StandardLogger implements the Logger interfaces using the standard library "log" package
type StandardLogger struct {
	// Logger may be provided optionally to customize the standard logger
	// If it is nil, the global default logger will be used
	Logger *log.Logger
}

// Log a line to the standard logger
func (l *StandardLogger) Log(msg string) {
	if l.Logger != nil {
		l.Logger.Println(msg)
	} else {
		log.Println(msg)
	}
}

// LogWriter implements the io.Writer interface by passing each line to the Logger after it hits a line break ('\n').
type LogWriter struct {
	logger Logger
	sb     *strings.Builder
}

// NewLogWriter creates a LogWriter using the logger provided
func NewLogWriter(logger Logger) *LogWriter {
	return &LogWriter{
		logger: logger,
		sb:     &strings.Builder{},
	}
}

// Writes to the string buffer, and passes each line to the logger after it reaches a new line
func (w *LogWriter) Write(p []byte) (n int, err error) {
	for _, c := range p {
		if c == '\n' { // Print log line
			w.logger.Log(w.sb.String())
			w.sb.Reset()
		} else {
			w.sb.WriteByte(c)
			n++
		}
	}
	return
}

// Flush a log line out and empty the buffer
func (w *LogWriter) Flush() error {
	w.logger.Log(w.sb.String())
	w.sb.Reset()
	return nil
}
