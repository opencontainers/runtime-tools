package main

import "github.com/opencontainers/runtime-tools/util/tap"

func main() {
	t := tap.New()
	t.Header(2)
	t.Ok(true, "first test")
	t.Pass("second test")
}
