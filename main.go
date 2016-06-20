package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "oci"
	app.Version = "0.0.1"
	app.Usage = "Utilities for OCI"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "error",
			Usage: "Log level (panic, fatal, error, warn, info, or debug)",
		},
	}

	app.Commands = []cli.Command{
		generateCommand,
		bundleValidateCommand,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func before(context *cli.Context) error {
	logLevelString := context.GlobalString("log-level")
	logLevel, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		logrus.Fatalf(err.Error())
	}
	logrus.SetLevel(logLevel)

	return nil
}
