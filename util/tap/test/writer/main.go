package main

import (
	"bytes"
	"os"

	"github.com/opencontainers/runtime-tools/util/tap"
)

func main() {
	buf := new(bytes.Buffer)
	t := tap.New()
	t.Writer = buf
	t.Header(2)
	t.Ok(true, "a test")
	t.Ok(buf.Len() > 0, "buffer has content")

	buf.WriteTo(os.Stdout)
}
