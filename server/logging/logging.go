package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	LevelChangedHook func(debug bool)
)

var (
	// The logging level.
	level = new(slog.LevelVar)
	// Callbacks for whenever we change the log level.
	levelChangedHooks []LevelChangedHook
)

// Setup slog defaults
func Setup(logfp string) io.Writer {
	multiw := io.MultiWriter(&lumberjack.Logger{
		Filename:   logfp,
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   false,
	}, os.Stdout)
	slog.SetDefault(slog.New(
		slog.NewTextHandler(multiw, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// `AddSource=true` adds source code location to each log,
				// this replaces the entire file path added with just it's
				// last dirname and filename.
				if a.Key == slog.SourceKey {
					s := a.Value.Any().(*slog.Source)
					s.File = filepath.Base(filepath.Dir(s.File)) +
						"/" + filepath.Base(s.File)
					return slog.Any(a.Key, s)
				}
				return a
			},
		}),
	))
	return multiw
}

// Add a hook to levelChangedHooks.
//
// **NOTE:** Ideally all hooks are added BEFORE the first `Level()` call, so
// all hooks are ran for the first init of the logger level.
func AddLevelChangedHook(lch LevelChangedHook) {
	levelChangedHooks = append(levelChangedHooks, lch)
}

// Set loggin level.
func Level(debug bool) {
	if debug {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo)
	}
	for i := range levelChangedHooks {
		levelChangedHooks[i](debug)
	}
	slog.Info("Logging level set", "level", level)
}
