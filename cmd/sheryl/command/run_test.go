package command

import (
	"bytes"
	"github.com/gari8/sheryl/pkg/config"
	"github.com/gari8/sheryl/pkg/log"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func Test_validate(t *testing.T) {
	cases := []struct {
		name    string
		arrange func() *Command
		assert  func(err error)
	}{
		{
			name: "should not be error",
			arrange: func() *Command {
				cmd := &Command{
					Cobra:  nil,
					Config: &config.Config{},
				}
				cmd.Config.Steps = []*config.Step{{Name: "test_1"}, {Name: "test_2"}, {Name: "test_3"}, {Name: "test_4"}}
				return cmd
			},
			assert: func(err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "should be error",
			arrange: func() *Command {
				cmd := &Command{
					Cobra:  nil,
					Config: &config.Config{},
				}
				cmd.Config.Steps = []*config.Step{{Name: "test_1"}, {Name: "test_2"}, {Name: "test_2"}}
				return cmd
			},
			assert: func(err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := c.arrange()
			c.assert(cmd.validate())
		})
	}
}

func Test_setLogHandler(t *testing.T) {
	cases := []struct {
		name    string
		arrange func() (format string, buf bytes.Buffer, check func(handler slog.Handler) bool)
		assert  func(flag bool)
	}{
		{
			name: "use simple log handler",
			arrange: func() (format string, buf bytes.Buffer, check func(handler slog.Handler) bool) {
				return OutputFormatSimple, *bytes.NewBuffer(nil), func(handler slog.Handler) bool {
					_, ok := handler.(*log.SimpleHandler)
					return ok
				}
			},
			assert: func(flag bool) {
				assert.True(t, flag)
			},
		},
		{
			name: "use json log handler",
			arrange: func() (format string, buf bytes.Buffer, check func(handler slog.Handler) bool) {
				return OutputFormatJSON, *bytes.NewBuffer(nil), func(handler slog.Handler) bool {
					_, ok := handler.(*slog.JSONHandler)
					return ok
				}
			},
			assert: func(flag bool) {
				assert.True(t, flag)
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			format, buf, check := c.arrange()
			setLogHandler(format, &buf)
			handler := slog.Default().Handler()
			c.assert(check(handler))
		})
	}
}
