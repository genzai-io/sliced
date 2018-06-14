package raft_service

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog"
)

type LoggerWriter struct {
	Logger zerolog.Logger
}

func (w *LoggerWriter) Write(buf []byte) (int, error) {
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
				w.Logger.Warn().Str("component", name).Msg(msg)
			case "DEBU", "DEBUG":
				w.Logger.Debug().Str("component", name).Msg(msg)
			case "INFO":
				w.Logger.Info().Str("component", name).Msg(msg)
			case "ERR", "ERRO", "ERROR":
				w.Logger.Error().Str("component", name).Msg(msg)

			default:
				w.Logger.Info().Str("component", name).Msg(msg)
			}
		} else {
			w.Logger.Info().Str("component", "raft").Msg(strings.TrimSpace(string(buf)))
		}
	} else {
		w.Logger.Info().Str("component", "raft").Msg(strings.TrimSpace(string(buf)))
	}
	return l, nil
}
