package main

import "github.com/mndrix/tap-go"

func main() {
	t := tap.New()
	t.Header(2)
	t.Ok(false, "first test")
	t.Fail("second test")
}
