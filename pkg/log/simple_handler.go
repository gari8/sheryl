package log

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log/slog"
	"strings"
)

const sep = "======================================================================================"

type SimpleHandler struct {
	w     io.Writer
	attrs []slog.Attr
}

func NewSimpleHandler(w io.Writer) *SimpleHandler {
	return &SimpleHandler{w, nil}
}
func (s *SimpleHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (s *SimpleHandler) Handle(_ context.Context, record slog.Record) error {
	var colorHandler func(format string, a ...any) string
	switch record.Level {
	case slog.LevelWarn:
		colorHandler = color.HiRedString
	case slog.LevelError:
		colorHandler = color.RedString
	case slog.LevelInfo:
		colorHandler = color.GreenString
	default:
		colorHandler = fmt.Sprintf
	}

	var logStrings []string
	record.Attrs(func(attr slog.Attr) bool {
		logStrings = append(logStrings, fmt.Sprintf("%s: %s", attr.Key, strings.Trim(attr.Value.String(), "\n")))
		return true
	})
	content := strings.Join(append([]string{record.Message}, logStrings...), " ")
	if len(logStrings) > 1 {
		content = fmt.Sprintf(
			"%s\n%s",
			record.Message,
			colorHandler(strings.Join(append([]string{sep}, append(logStrings, sep)...), "\n")),
		)
	}
	_, err := fmt.Fprintln(s.w, content)
	return err
}
func (s *SimpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler { s.attrs = attrs; return s }
func (s *SimpleHandler) WithGroup(name string) slog.Handler       { return s }
