package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/flashbots/prometheus-sns-lambda-webhook/config"
	"github.com/flashbots/prometheus-sns-lambda-webhook/logutils"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var (
	version = "development"
)

var (
	ErrLoggingFailedToSetup = errors.New("failed to setup logging")
)

func main() {
	cfg := &config.Config{}

	flagLogLevel := &cli.StringFlag{
		Destination: &cfg.Log.Level,
		EnvVars:     []string{"LOG_LEVEL"},
		Name:        "log-level",
		Usage:       "logging level",
		Value:       "info",
	}

	flagLogMode := &cli.StringFlag{
		Destination: &cfg.Log.Mode,
		EnvVars:     []string{"LOG_MODE"},
		Name:        "log-mode",
		Usage:       "logging mode",
		Value:       "prod",
	}

	if version == "development" {
		flagLogLevel.Value = "debug"
		flagLogMode.Value = "dev"
	}

	app := &cli.App{
		Name:    "sns-lambda-webhook",
		Usage:   "Receive prometheus alerts via SNS and push them to a webhook",
		Version: version,

		Flags: []cli.Flag{
			flagLogLevel,
			flagLogMode,
		},

		Before: func(_ *cli.Context) error {
			l, err := logutils.NewLogger(&cfg.Log)
			if err != nil {
				return fmt.Errorf("%w: %w",
					ErrLoggingFailedToSetup, err,
				)
			}
			zap.ReplaceGlobals(l)
			return nil
		},

		DefaultCommand: "lambda",

		Commands: []*cli.Command{
			CommandLambda(cfg),
		},
	}

	defer func() {
		zap.L().Sync() //nolint:errcheck
	}()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed with error:\n\n%s\n\n", err.Error())
		os.Exit(1)
	}
}
