package core

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog"
)

type raftLoggerWriter struct {
	logger zerolog.Logger
}

func (w *raftLoggerWriter) Write(buf []byte) (int, error) {
	l := len(buf)
	b := buf
	lidx := bytes.IndexByte(b, '[')
	if lidx > -1 {
		b = b[lidx+1:]
		idx := bytes.IndexByte(b, ']')
		if idx > 0 {
			level := string(b[0:idx])

			b = b[idx+1:]
			name := "raft"
			idx = bytes.IndexByte(b, ':')

			if idx > 0 {
				name = string(bytes.TrimSpace(b[:idx]))
				b = b[idx+1:]
			}

			msg := strings.TrimSpace(string(b))
			switch level {
			case "WARN":
				w.logger.Warn().Str("component", name).Msg(msg)
			case "DEBU", "DEBUG":
				w.logger.Debug().Str("component", name).Msg(msg)
			case "INFO":
				w.logger.Info().Str("component", name).Msg(msg)
			case "ERR", "ERRO", "ERROR":
				w.logger.Error().Str("component", name).Msg(msg)

			default:
				w.logger.Info().Str("component", name).Msg(msg)
			}
		} else {
			w.logger.Info().Str("component", "raft").Msg(strings.TrimSpace(string(buf)))
		}
	} else {
		w.logger.Info().Str("component", "raft").Msg(strings.TrimSpace(string(buf)))
	}
	return l, nil
}
