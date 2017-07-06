package main

import (
	"fmt"
	"strings"

	"github.com/opencontainers/runtime-tools/validate"
	"github.com/urfave/cli"
)

var bundleValidateFlags = []cli.Flag{
	cli.StringFlag{Name: "path", Value: ".", Usage: "path to a bundle"},
	cli.StringFlag{Name: "platform", Value: "linux", Usage: "platform of the target bundle (linux, windows, solaris)"},
}

var bundleValidateCommand = cli.Command{
	Name:   "validate",
	Usage:  "validate an OCI bundle",
	Flags:  bundleValidateFlags,
	Before: before,
	Action: func(context *cli.Context) error {
		hostSpecific := context.GlobalBool("host-specific")
		inputPath := context.String("path")
		platform := context.String("platform")
		v, err := validate.NewValidatorFromPath(inputPath, hostSpecific, platform)
		if err != nil {
			return err
		}

		errMsgs := v.CheckAll()
		if len(errMsgs) > 0 {
			return fmt.Errorf("%d Errors detected:\n%s", len(errMsgs), strings.Join(errMsgs, "\n"))

		}
		fmt.Println("Bundle validation succeeded.")
		return nil
	},
}
