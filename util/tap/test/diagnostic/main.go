package main

import "github.com/mndrix/tap-go"

func main() {
	t := tap.New()
	t.Header(1)
	t.Diagnostic("expecting all to be well")
	t.Diagnosticf("here's some perfectly magical output: %d %s 0x%X.", 6, "abracadabra", 28)
	t.Pass("all good")
}
