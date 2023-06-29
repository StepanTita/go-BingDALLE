package cli

import (
	"runtime/debug"

	"github.com/urfave/cli/v2"

	"github.com/sirupsen/logrus"

	"github.com/StepanTita/go-BingDALLE/config"
	"github.com/StepanTita/go-BingDALLE/services/communicator"
)

func Run(args []string) bool {
	var cliConfig config.CliConfig

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Log level ['debug', 'info', 'warn', 'error', 'fatal']",
				Value:       "info",
				Category:    "Miscellaneous:",
				Destination: &cliConfig.LogLevel,
			},
			&cli.StringFlag{
				Name:        "bing-url",
				Usage:       "Bing URL (e.g. https://www.bing.com)",
				Category:    "Networking:",
				Required:    false,
				Value:       "https://www.bing.com",
				Destination: &cliConfig.ApiUrl,
			},
			&cli.StringFlag{
				Name:        "proxy",
				Usage:       "Proxy URL (e.g. https://127.0.0.1:1080)",
				Category:    "Networking:",
				Required:    false,
				Destination: &cliConfig.Proxy,
			},
			&cli.StringFlag{
				Name:        "u-auth-cookie",
				Usage:       "Cookie value to authenticate request",
				Category:    "Miscellaneous:",
				Required:    true,
				Destination: &cliConfig.UCookie,
			},
		},
		Commands: cli.Commands{
			{
				Name:  "run",
				Usage: "run DALLE daemon",
				Action: func(c *cli.Context) error {
					cfg := config.NewFromCLI(cliConfig)
					log := cfg.Logging()

					log.WithField("version", cfg.Version()).Info("Running version...")

					defer func() {
						if rvr := recover(); rvr != nil {
							log.Error("internal panicked: ", rvr, string(debug.Stack()))
						}
					}()

					comm := communicator.New(cfg)

					return comm.Run(c.Context)
				},
			},
		},
	}

	if err := app.Run(args); err != nil {
		logrus.Error(err, ": service initialization failed")
		return false
	}

	return true
}
