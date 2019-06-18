package proxy

import (
	"strings"
)

type logWriter struct {
	logger Logger
	sb     *strings.Builder
}

func newLogWriter(logger Logger) *logWriter {
	return &logWriter{
		logger: logger,
		sb:     &strings.Builder{},
	}
}

func (w *logWriter) Write(p []byte) (n int, err error) {
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
