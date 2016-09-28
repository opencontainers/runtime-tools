package tap

import "fmt"
import "testing/quick"

type T struct {
	nextTestNumber int
}

// New creates a new Tap value
func New() *T {
	return &T{
		nextTestNumber: 1,
	}
}

// Header displays a TAP header including version number and expected
// number of tests to run.  For an unknown number of tests, set
// testCount to zero (in which case the plan is not written); this is
// useful with AutoPlan.
func (t *T) Header(testCount int) {
	fmt.Printf("TAP version 13\n")
	if testCount > 0 {
		fmt.Printf("1..%d\n", testCount)
	}
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

// Check runs randomized tests against a function just as "testing/quick.Check"
// does.  Success or failure generate appropriate TAP output.
func (t *T) Check(function interface{}, description string) {
	err := quick.Check(function, nil)
	if err == nil {
		t.Ok(true, description)
		return
	}

	fmt.Printf("# %s\n", err)
	t.Ok(false, description)
}

// Count returns the number of tests completed so far.
func (t *T) Count() int {
	return t.nextTestNumber - 1
}

// AutoPlan generates a test plan based on the number of tests that were run.
func (t *T) AutoPlan() {
	fmt.Printf("1..%d\n", t.nextTestNumber-1)
}
