package main

import (
	"os"

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
			Name:   "environment-name",
			Usage:  "environment name in the app to update",
			EnvVar: "PLUGIN_ENVIRONMENT_NAME",
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
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func run(c *cli.Context) error {
	plugin := Plugin{
		Key:             c.String("access-key"),
		Secret:          c.String("secret-key"),
		Application:     c.String("application"),
		EnvironmentName: c.String("environment-name"),
		VersionLabel:    c.String("version-label"),
		Region:          c.String("region"),
		Timeout:         c.String("timeout"),
	}

	return plugin.Exec()
}
