package main

import (
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

var version string

func main() {
	app := cli.NewApp()
	app.Name = "Beanstalk deployment checker plugin"
	app.Usage = "beanstalk deployment checker plugin"
	app.Action = run
	app.Version = version
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "access-key",
			Usage:  "aws access key",
			EnvVar: "PLUGIN_ACCESS_KEY,AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "secret-key",
			Usage:  "aws secret key",
			EnvVar: "PLUGIN_SECRET_KEY,AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "application",
			Usage:  "application name for beanstalk",
			EnvVar: "PLUGIN_APPLICATION",
		},
		cli.StringFlag{
			Name:   "environment",
			Usage:  "optional environment name for beanstalk",
			EnvVar: "PLUGIN_ENVIRONMENT",
		},
		cli.StringFlag{
			Name:   "version-label",
			Usage:  "version label for the app",
			EnvVar: "PLUGIN_VERSION_LABEL",
		},
		cli.StringFlag{
			Name:   "region",
			Usage:  "aws region",
			Value:  "us-east-1",
			EnvVar: "PLUGIN_REGION",
		},
		cli.StringFlag{
			Name:   "timeout",
			Usage:  "deploy timeout in minutes",
			Value:  "30",
			EnvVar: "PLUGIN_TIMEOUT",
		},
		cli.StringFlag{
			Name:   "tick",
			Usage:  "deploy tick in seconds",
			Value:  "20",
			EnvVar: "PLUGIN_TICK",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "set to true for debug log",
			EnvVar: "PLUGIN_DEBUG",
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	timeout, err := strconv.Atoi(c.String("timeout"))

	if err != nil {
		log.WithFields(log.Fields{
			"timeout": c.String("timeout"),
			"error":   err,
		}).Error("invalid timeout configuration")
		return err
	}

	tick, err := strconv.Atoi(c.String("tick"))

	if err != nil {
		log.WithFields(log.Fields{
			"tick":  c.String("tick"),
			"error": err,
		}).Error("invalid tick configuration")
		return err
	}

	plugin := Plugin{
		Key:          c.String("access-key"),
		Secret:       c.String("secret-key"),
		Application:  c.String("application"),
		Environment:  c.String("environment"),
		VersionLabel: c.String("version-label"),
		Region:       c.String("region"),
		Debug:        c.Bool("debug"),
		Tick:         time.Duration(tick) * time.Second,
		Timeout:      time.Duration(timeout) * time.Minute,
	}

	return plugin.Exec()
}
