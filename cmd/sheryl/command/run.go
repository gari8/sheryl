package command

import (
	"cmp"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/gari8/sheryl/pkg/config"
	"github.com/gari8/sheryl/pkg/log"
	"github.com/gari8/sheryl/pkg/types"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
)

const (
	runCommand = "run"

	// OutputFormatSimple simple log format with output flag
	OutputFormatSimple = "simple"
	// OutputFormatJSON json log format with output flag
	OutputFormatJSON = "json"
)

func (c *Command) cmdRun() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:           runCommand,
		Short:         "run sheryl",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       c.prepare,
		Run: func(cmd *cobra.Command, args []string) {
			// enable the verbose flag to get detailed logs
			verbose, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				slog.ErrorContext(cmd.Context(), err.Error())
				os.Exit(1)
			}
			steps, err := c.Config.ToTypesSteps()
			if err != nil {
				slog.ErrorContext(cmd.Context(), err.Error())
				os.Exit(1)
			}
			beforeSteps := make(map[string]*types.Step)
			for _, step := range steps {
				step.Verbose = verbose
				// add a retry setting in case of failure
				retries := step.Retries
				output, err := retry.DoWithData(func() ([]byte, error) {
					output, err := step.Run(cmd.Context(), beforeSteps)
					if err != nil {
						retries--
						slog.ErrorContext(cmd.Context(), fmt.Errorf("【%s】is failed [%w]", step.Name, err).Error(), slog.Any("pid", step.PID))
					}
					return output, err
				}, retry.Delay(step.Interval), retry.RetryIf(func(_ error) bool {
					return retries > 0
				}))
				if err != nil {
					slog.ErrorContext(cmd.Context(), fmt.Errorf("【%s】is failed [%w]", step.Name, err).Error(), step.Attributes()...)
					os.Exit(1)
				} else {
					if output != nil {
						step.Output = fmt.Sprintf("%s", string(output))
					}
					slog.InfoContext(cmd.Context(), fmt.Sprintf("【%s】is success", step.Name), step.Attributes()...)
				}
				if _, exists := beforeSteps[step.Name]; !exists {
					beforeSteps[step.Name] = step
				}
			}
		},
	}
	cmd.Flags().BoolP("verbose", "v", false, "verbose")
	cmd.Flags().StringP("config", "c", "", "config path")
	return
}

func (c *Command) prepare(cmd *cobra.Command, args []string) error {
	outputFormat, err := cmd.Flags().GetString("output")
	setLogHandler(outputFormat, os.Stdout)
	if err != nil {
		return err
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	configPath = cmp.Or(configPath, "./")
	c.Config, err = config.Load(configPath)
	if err != nil {
		return err
	}
	return c.validate()
}

func (c *Command) validate() error {
	stepNameStore := make(map[string]struct{})
	for _, s := range c.Config.Steps {
		if _, exists := stepNameStore[s.Name]; exists {
			return fmt.Errorf("step name %s is duplicated", s.Name)
		}
		stepNameStore[s.Name] = struct{}{}
	}
	return nil
}

func setLogHandler(format string, w io.Writer) {
	var logHandler slog.Handler
	switch format {
	case OutputFormatJSON:
		logHandler = slog.NewJSONHandler(w, nil)
	case OutputFormatSimple:
		logHandler = log.NewSimpleHandler(w)
	default:
		logHandler = log.NewSimpleHandler(w)
	}
	slog.SetDefault(slog.New(logHandler))
}
