package tap

import "fmt"

type T struct {
	nextTestNumber int
}

// New creates a new Tap value
func New() *T {
	return &T{}
}

// Header displays a TAP header including version number and expected
// number of tests to run.
func (t *T) Header(testCount int) {
	fmt.Printf("TAP version 13\n")
	fmt.Printf("1..%d\n", testCount)
}

// Ok generates TAP output indicating whether a test passed or failed.
func (t *T) Ok(test bool, description string) {
	// did the test pass or not?
	ok := "ok"
	if !test {
		ok = "not ok"
	}

	fmt.Printf("%s %d - %s\n", ok, t.nextTestNumber, description)
	t.nextTestNumber++
}
