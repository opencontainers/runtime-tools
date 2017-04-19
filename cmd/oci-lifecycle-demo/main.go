package main

import (
	"fmt"

	l "github.com/opencontainers/runtime-tools/lifecycle"
)

func TestDemo() {
	runtime := "runc"

	type ActionCase struct {
		action   l.LifecycleAction
		args     []string
		expected bool
	}
	cases := []struct {
		bundle  string
		id      string
		actions []ActionCase
	}{
		{"noexist_bundle", "1", []ActionCase{
			{action: l.LifecycleCreate, expected: false}},
		},
		{"exist_bundle", "1", []ActionCase{
			{action: l.LifecycleCreate, expected: true}},
		},
		{"exist_bundle", "1", []ActionCase{
			{action: l.LifecycleCreate, expected: true},
			{action: l.LifecycleCreate, expected: false}},
		},
		{"exist_bundle_run_err", "1", []ActionCase{
			{action: l.LifecycleCreate, expected: true},
			{action: l.LifecycleStart, expected: false}},
		},
	}
	for _, c := range cases {
		//TODO: Prepare c.bundle
		lc, _ := l.NewLifecycle(runtime, c.bundle, c.id)
		for _, a := range c.actions {
			_, err := lc.Operate(a.action)
			if err != nil && a.expected == true {
				fmt.Printf("Failed in test")
				break
			} else if err == nil && a.expected == false {
				fmt.Printf("Failed in test")
				break
			}
		}

		//TODO: Remove it anyway
		lc.Operate(l.LifecycleDelete)
	}
}

func main() {
	runtime := "runc"
	bundle := "bundle"
	id := "25"
	lc, _ := l.NewLifecycle(runtime, bundle, id)

	fmt.Println("create")
	out, err := lc.Operate(l.LifecycleCreate)
	fmt.Println(string(out), err)

	fmt.Println("state")
	out, err = lc.Operate(l.LifecycleState)
	fmt.Println(string(out), err)

	fmt.Println("start")
	out, err = lc.Operate(l.LifecycleStart)
	fmt.Println(string(out), err)

	fmt.Println("state")
	out, _ = lc.Operate(l.LifecycleState)
	fmt.Println(string(out))

	fmt.Println("delete")
	out, _ = lc.Operate(l.LifecycleDelete)
	fmt.Println(string(out))

	fmt.Println("state")
	out, _ = lc.Operate(l.LifecycleState)
	fmt.Println(string(out))
}
