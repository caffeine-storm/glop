package glog

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type handlerAtLevel struct {
	slog.Handler
	lvl slog.Level
}

func (hdl *handlerAtLevel) Enabled(ctx context.Context, lvl slog.Level) bool {
	return lvl >= hdl.lvl
}

func Relevel(in *slog.Logger, lvl slog.Level) *slog.Logger {
	handler := in.Handler()

	handler = &handlerAtLevel{
		Handler: handler, lvl: lvl,
	}

	return slog.New(handler)
}

type Opts struct {
	Output io.Writer
	Level  slog.Leveler
}

func New(options *Opts) *slog.Logger {
	if options == nil {
		options = &Opts{}
	}
	if options.Output == nil {
		options.Output = os.Stderr
	}
	if options.Level == nil {
		options.Level = slog.LevelInfo
	}

	slogopts := &slog.HandlerOptions{
		Level: options.Level,
	}
	return slog.New(slog.NewTextHandler(options.Output, slogopts))
}

func DebugLogger() *slog.Logger {
	return New(&Opts{
		Level: slog.LevelDebug,
	})
}
