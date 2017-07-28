package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// gitCommit will be the hash that the binary was built from
// and will be populated by the Makefile
var gitCommit = ""

func main() {
	app := cli.NewApp()
	app.Name = "oci-runtime-tool"
	if gitCommit != "" {
		app.Version = fmt.Sprintf("0.0.1, commit: %s", gitCommit)
	} else {
		app.Version = "0.0.1"
	}
	app.Usage = "OCI (Open Container Initiative) runtime tools"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "host-specific",
			Usage: "generate host-specific configs or do host-specific validations",
		},
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
