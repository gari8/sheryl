package command

import (
	"github.com/gari8/sheryl/pkg/config"
	"github.com/spf13/cobra"
)

type Command struct {
	Cobra  *cobra.Command
	Config *config.Config
}

func New(version string) *Command {
	c := &Command{
		Cobra: &cobra.Command{
			Use:     "sheryl",
			Short:   "Sheryl is a tool that integrates multiple shell scripts based on your rules.",
			Version: version,
		},
	}
	c.Cobra.AddCommand(
		c.cmdRun(),
	)
	c.Cobra.PersistentFlags().StringP("output", "o", "simple", "select output format with [\"simple\", \"json\"]")
	return c
}
