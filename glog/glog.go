package glog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type handlerAtLevel struct {
	slog.Handler
	lvl slog.Level
}

func (hdl *handlerAtLevel) Enabled(ctx context.Context, lvl slog.Level) bool {
	return lvl >= hdl.lvl
}

func Relevel(in Logger, lvl slog.Level) Logger {
	handler := in.Handler()

	handler = &handlerAtLevel{
		Handler: handler, lvl: lvl,
	}

	return &traceLogger{
		Logger: slog.New(handler),
	}
}

func WithRedirect(in Logger, out io.Writer) Logger {
	opts := in.GetOpts()

	wrapped := slog.New(slog.NewTextHandler(out, &opts))
	return &traceLogger{
		Logger: wrapped,
	}
}

type Opts struct {
	Output         io.Writer
	Level          slog.Leveler
	DoNotAddSource bool
}

func trimLeadingDir(in string) string {
	// the 'source' attr gets lots of leading dirs; let's trim away anything
	// above 'glop'.
	parts := strings.Split(in, string(filepath.Separator))
	trimpoints := map[string]bool{
		"glop":   true,
		"haunts": true,
	}

	for i := len(parts); i > 0; {
		i--
		if trimpoints[parts[i]] {
			return path.Join(parts[i+1:]...)
		}
	}

	slog.Default().Warn("glog.trimLeadingDir: no trim point found", "input", in, "trimpoints", trimpoints)

	return in
}

func trimLeadingDirNoise(groups []string, a slog.Attr) slog.Attr {
	// Let Attrs that aren't "source" pass through.
	if a.Key != "source" {
		return a
	}

	trimmed := trimLeadingDir(a.Value.String())
	lastSpace := strings.LastIndex(trimmed, " ")
	fileColonLine := trimmed[0:lastSpace] + ":" + trimmed[lastSpace+1:]
	return slog.String("source", fileColonLine)
}

func New(options *Opts) Logger {
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
		AddSource:   !options.DoNotAddSource,
		Level:       options.Level,
		ReplaceAttr: trimLeadingDirNoise,
	}
	wrapped := slog.New(slog.NewTextHandler(options.Output, slogopts))
	return &traceLogger{
		Logger:         wrapped,
		handlerOptions: *slogopts,
	}
}

// Annoyingly, slog and log use structs instead of interfaces... we'll make our
// own!
type Slogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Handler() slog.Handler
	Log(ctx context.Context, lvl slog.Level, msg string, args ...interface{})
	Enabled(ctx context.Context, lvl slog.Level) bool
}

type Logger interface {
	Slogger
	Trace(msg string, args ...interface{})
	GetOpts() slog.HandlerOptions
}

type traceLogger struct {
	*slog.Logger
	handlerOptions slog.HandlerOptions
}

var _ Logger = (*traceLogger)(nil)

const LevelTrace = slog.LevelDebug - 4

func (log *traceLogger) Trace(msg string, args ...interface{}) {
	log.Log(context.Background(), LevelTrace, msg, args...)
}

func (log *traceLogger) GetOpts() slog.HandlerOptions {
	return log.handlerOptions
}

func TraceLogger() Logger {
	return New(&Opts{
		Level: slog.LevelInfo, // Trace calls will be ignored unless the caller re-levels
	})
}

func InfoLogger() Logger {
	return New(&Opts{
		Level: slog.LevelInfo,
	})
}

func DebugLogger() Logger {
	return New(&Opts{
		// TODO(tmckee): we should use LevelInfo and expect callers who _really_
		// want the message to re-level.
		// Even better, have a 'current level' in glog that we read from on
		// construction.
		Level: slog.LevelDebug,
	})
}

func WarningLogger() Logger {
	return New(&Opts{
		Level: slog.LevelWarn,
	})
}

func ErrorLogger() Logger {
	return New(&Opts{
		Level: slog.LevelError,
	})
}

func VoidLogger() Logger {
	return New(&Opts{
		Output: io.Discard,
	})
}
