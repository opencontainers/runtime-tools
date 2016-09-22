package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/sanitize"
	"github.com/urfave/cli"
)

var sanitizeFlags = []cli.Flag{
	cli.StringFlag{Name: "output", Usage: "output file (defaults to stdout)"},
}

var sanitizeCommand = cli.Command{
	Name:   "sanitize",
	Usage:  "sanitize an OCI runtime configuration file",
	Flags:  sanitizeFlags,
	Before: before,
	Action: func(context *cli.Context) (err error) {
		var reader io.ReadCloser
		if context.NArg() == 0 {
			reader = os.Stdin
		} else if context.NArg() == 1 {
			reader, err = os.Open(context.Args().First())
			if err != nil {
				return err
			}
			defer reader.Close()
		} else {
			return fmt.Errorf("too many arguments (%d > 1)", context.NArg())
		}

		var config rspec.Spec
		err = json.NewDecoder(reader).Decode(&config)
		if err != nil {
			return err
		}

		err = reader.Close()
		if err != nil {
			return err
		}

		err = sanitize.Sanitize(&config)
		if err != nil {
			return err
		}

		var writer io.WriteCloser
		if context.IsSet("output") {
			writer, err = os.OpenFile(context.String("output"), os.O_WRONLY | os.O_TRUNC, 0)
			if err != nil {
				return err
			}
			defer writer.Close()
		} else {
			writer = os.Stdout
		}

		encoder := json.NewEncoder(writer)
		return encoder.Encode(&config)
	},
}
