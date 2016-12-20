package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/lifecycle"
)

func main() {
	runtime := "runc"
	bundle := "bundle"
	id := "25"
	lc, _ := lifecycle.NewLifecycle(runtime, bundle, id)

	fmt.Println("create")
	out, err := lc.Operate(lifecycle.LifecycleCreate)
	fmt.Println(string(out), err)

	fmt.Println("state")
	out, err = lc.Operate(lifecycle.LifecycleState)
	fmt.Println(string(out), err)

	fmt.Println("start")
	out, err = lc.Operate(lifecycle.LifecycleStart)
	fmt.Println(string(out), err)

	fmt.Println("state")
	out, _ = lc.Operate(lifecycle.LifecycleState)
	fmt.Println(string(out))

	fmt.Println("delete")
	out, _ = lc.Operate(lifecycle.LifecycleDelete)
	fmt.Println(string(out))

	fmt.Println("state")
	out, _ = lc.Operate(lifecycle.LifecycleState)
	fmt.Println(string(out))
}
