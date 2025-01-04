package main

import (
	"github.com/gari8/sheryl/cmd/sheryl/command"
	"log/slog"
	"os"
)

func main() {
	cmd := command.New(version).Cobra
	if err := cmd.Execute(); err != nil {
		slog.ErrorContext(cmd.Context(), err.Error())
		os.Exit(1)
	}
}
