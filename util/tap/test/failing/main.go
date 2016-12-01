package main

import "github.com/mndrix/tap-go"

func main() {
	t := tap.New()
	t.Header(1)
	t.Ok(false, "first test")
}
