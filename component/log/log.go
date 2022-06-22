package log

import (
	"context"
	"os"
	"path"
	"runtime/debug"
	"strconv"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = "2006/01/02 15:04:05"
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		fileName := path.Base(file)
		dir := path.Dir(file)
		lastDir := path.Base(dir)
		return path.Join(lastDir, fileName) + ":" + strconv.Itoa(line)
	}

}

type CtxLogger interface {
	// debug 级别
	Print(ctx context.Context, msg string, kvs ...interface{})

	// debug 级别
	Debug(ctx context.Context, msg string, kvs ...interface{})

	// info 级别
	Info(ctx context.Context, msg string, kvs ...interface{})

	// warning 级别
	Warning(ctx context.Context, msg string, kvs ...interface{})

	// error 级别
	Error(ctx context.Context, msg string, kvs ...interface{})

	// panic 级别
	Panic(ctx context.Context, msg string, kvs ...interface{})

	// fatal 级别
	Fatal(ctx context.Context, msg string, kvs ...interface{})
}

type Logger struct {
	w zerolog.Logger
}

func New(l zerolog.Logger) *Logger {
	return &Logger{w: l}
}

func (l *Logger) Print(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Debug().Msg(msg)
}

func (l *Logger) Debug(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Debug().Msg(msg)
}

func (l *Logger) Info(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Info().Msg(msg)
}

func (l *Logger) Warning(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Warn().Msg(msg)
}

func (l *Logger) Error(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Error().Msg(msg)
}

func (l *Logger) Panic(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Panic().Msg(msg)

}

func (l *Logger) Fatal(ctx context.Context, msg string, kvs ...interface{}) {
	lg := l.w.With().Fields(ctxFields(ctx, kvs)).Logger()
	lg.Fatal().Msg(msg)
}

// frame 4 global method 调用
// frame 3 global our logger method 调用
// frame 2 global zerolog method 调用
// frame 1 global zerolog event msg调用
// frame 0 global zerolog event write调用
var std CtxLogger = New(zerolog.New(os.Stdout).With().CallerWithSkipFrameCount(4).Timestamp().Logger())

// Default returns the standard logger used by the package-level output functions.
func Default() CtxLogger { return std }

func Print(ctx context.Context, msg string, kvs ...interface{}) {
	std.Print(ctx, msg, kvs...)
}

func Debug(ctx context.Context, msg string, kvs ...interface{}) {
	std.Debug(ctx, msg, kvs...)
}

func Info(ctx context.Context, msg string, kvs ...interface{}) {
	std.Info(ctx, msg, kvs...)
}

func Warning(ctx context.Context, msg string, kvs ...interface{}) {
	std.Warning(ctx, msg, kvs...)
}

func Error(ctx context.Context, msg string, kvs ...interface{}) {
	std.Error(ctx, msg, kvs...)
}

func Fatal(ctx context.Context, msg string, kvs ...interface{}) {
	std.Fatal(ctx, msg, kvs...)
}

func Panic(ctx context.Context, msg string, kvs ...interface{}) {
	std.Panic(ctx, msg, kvs...)
}

func Stack(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append(kvs, "stacktrace", debug.Stack())
	std.Error(ctx, msg, kvs...)
}

func Fatalln(kvs ...interface{}) {
	std.Fatal(context.Background(), "", kvs...)
}

func Println(kvs ...interface{}) {
	std.Print(context.Background(), "", kvs...)
}
