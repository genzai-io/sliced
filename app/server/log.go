package server

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/slice-d/genzai"
	"github.com/rs/zerolog"
)

type zeroLogger struct {
}

func (l *zeroLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	logger := moved.Logger.With().Str("logger", "http")

	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		logger = logger.Str("request_id", reqID)
	}
	logger = logger.Str("method", r.Method)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logger = logger.Str("scheme", scheme)

	logger = logger.Str("remote_addr", r.RemoteAddr)
	//logger = logger.Str("host", r.Host)
	logger = logger.Str("uri", r.RequestURI)
	logger = logger.Str("proto", r.Proto)

	return &logEntry{
		request: r,
		logger:  &logger,
		buf:     &bytes.Buffer{},
	}
}

type logEntry struct {
	request *http.Request
	logger  *zerolog.Context
	buf     *bytes.Buffer
}

func (l *logEntry) Write(status, bytes int, elapsed time.Duration) {
	logger := l.logger.Int("status", status).Dur("elapsed", elapsed).Int("bytes", bytes).Logger()
	logger.Info().Msg("")
}

func (l *logEntry) Panic(v interface{}, stack []byte) {
	logger := l.logger.Str("stack", string(stack)).Logger()
	logger.Error().Msgf("%s", v)
}
