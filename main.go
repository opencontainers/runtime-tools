package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "oci"
	app.Version = "0.0.1"
	app.Usage = "Utilities for OCI"

	app.Commands = []cli.Command{
		generateCommand,
		bundleValidateCommand,
		validateCommand,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
