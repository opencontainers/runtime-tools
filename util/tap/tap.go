package tap

import "fmt"

// Header displays a TAP header including version number and expected
// number of tests to run.
func Header(testCount int) {
	fmt.Printf("TAP version 13\n")
	fmt.Printf("1..%d\n", testCount)
}

// Ok generates TAP output indicating that a test has passed
func Ok(testNumber int, description string) {
	fmt.Printf("ok %d - %s\n", testNumber, description)
}

// NotOk generates TAP output indicating that a test has failed
func NotOk(testNumber int, description string) {
	fmt.Printf("not ok %d - %s\n", testNumber, description)
}
