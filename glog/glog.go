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

func Relevel(in *slog.Logger, lvl slog.Level) *slog.Logger {
	handler := in.Handler()

	handler = &handlerAtLevel{
		Handler: handler, lvl: lvl,
	}

	return slog.New(handler)
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
		"glop": true,
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
		AddSource:   !options.DoNotAddSource,
		Level:       options.Level,
		ReplaceAttr: trimLeadingDirNoise,
	}
	return slog.New(slog.NewTextHandler(options.Output, slogopts))
}

func DebugLogger() *slog.Logger {
	return New(&Opts{
		Level: slog.LevelDebug,
	})
}

func VoidLogger() *slog.Logger {
	return New(&Opts{
		Output: io.Discard,
	})
}
