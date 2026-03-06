package main

import (
	"fmt"
	"os"
)

func main() {
	sub := "serve"
	args := os.Args[1:]

	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		sub = args[0]
		args = args[1:]
	}

	switch sub {
	case "serve":
		runServe(args)
	case "provision":
		runProvision(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\nUsage: gpuctl [serve|provision] [flags]\n", sub)
		os.Exit(1)
	}
}
