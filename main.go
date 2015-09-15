package main

import (
	"os"

	"github.com/mrunalp/ocitools/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/mrunalp/ocitools/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "oci"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		generateCommand,
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
