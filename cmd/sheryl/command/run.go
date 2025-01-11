package command

import (
	"cmp"
	"context"
	"fmt"
	"github.com/avast/retry-go/v4"
	"github.com/gari8/sheryl/pkg/config"
	"github.com/gari8/sheryl/pkg/log"
	"github.com/gari8/sheryl/pkg/types"
	"github.com/spf13/cobra"
	"io"
	"log/slog"
	"os"
	"time"
)

const (
	runCommand           = "run"
	prepareCtxObjectName = "prepareCtxObjectName"

	// OutputFormatSimple simple log format with output flag
	OutputFormatSimple = "simple"
	// OutputFormatJSON json log format with output flag
	OutputFormatJSON = "json"
)

type prepareRunCtxObject struct {
	Verbose bool
	Steps   []*types.Step
}

func newPCO(
	verbose bool,
	steps []*types.Step,
) *prepareRunCtxObject {
	return &prepareRunCtxObject{
		Verbose: verbose,
		Steps:   steps,
	}
}

func (p *prepareRunCtxObject) marshalCtx(parentCtx context.Context) context.Context {
	return context.WithValue(parentCtx, prepareCtxObjectName, *p)
}

func unmarshalCtx[T any](ctx context.Context, keyName string, out *T) error {
	val := ctx.Value(keyName)
	res, ok := val.(T)
	if !ok {
		return fmt.Errorf("%s is not found", keyName)
	}
	*out = res
	return nil
}

func (c *Command) cmdRun() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:           runCommand,
		Short:         "run sheryl",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       c.prepareRun,
		Run: func(cmd *cobra.Command, args []string) {
			var pco prepareRunCtxObject
			if err := unmarshalCtx(cmd.Context(), prepareCtxObjectName, &pco); err != nil {
				slog.ErrorContext(cmd.Context(), fmt.Errorf("context unmarshal failed: %w", err).Error())
			}
			beforeSteps := make(map[string]*types.Step)
			for _, step := range pco.Steps {
				step.Verbose = pco.Verbose
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
					slog.ErrorContext(cmd.Context(), fmt.Errorf("【%s】is failed [%w]", step.Name, err).Error(), log.NewAttr(step).Add()...)
					os.Exit(1)
				} else {
					if output != nil {
						step.Output = fmt.Sprintf("%s", string(output))
					}
					slog.InfoContext(cmd.Context(), fmt.Sprintf("【%s】is success", step.Name), log.NewAttr(step).Add()...)
				}
				if _, exists := beforeSteps[step.Name]; !exists {
					beforeSteps[step.Name] = step
				}
			}
			slog.InfoContext(cmd.Context(), "summary", log.NewAttr(newSummary(pco.Steps)).Add()...)
		},
	}
	cmd.Flags().BoolP("verbose", "v", false, "verbose")
	cmd.Flags().StringP("config", "c", "", "config path")
	return
}

func (c *Command) prepareRun(cmd *cobra.Command, _ []string) error {
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
	if c.Config, err = config.Load(configPath); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}
	// enable the verbose flag to get detailed logs
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	steps, err := c.Config.ToTypesSteps()
	if err != nil {
		return err
	}
	cmd.SetContext(newPCO(verbose, steps).marshalCtx(cmd.Context()))
	return nil
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

type execResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type summary struct {
	Results      []execResult  `json:"results"`
	ExecTime     time.Duration `json:"execTime"`
	SuccessCount string        `json:"success"`
}

func newSummary(steps []*types.Step) *summary {
	var res []execResult
	var successes int
	var sumDuration time.Duration
	for _, step := range steps {
		getStatus := func(success bool) string {
			if success {
				return "success"
			} else {
				return "failed"
			}
		}
		res = append(res, execResult{
			Name:   step.Name,
			Status: getStatus(!step.Failed),
		})
		if !step.Failed {
			successes += 1
		}
		sumDuration += step.Duration
	}
	return &summary{
		Results:      res,
		ExecTime:     sumDuration,
		SuccessCount: fmt.Sprintf("%d/%d", successes, len(steps)),
	}
}
