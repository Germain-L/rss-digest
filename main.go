package main

import (
	"fmt"
	"os"
)

func main() {
	switch {
	case len(os.Args) > 1 && os.Args[1] == "--serve":
		startServer()
	case len(os.Args) > 1 && os.Args[1] == "--generate":
		runGenerate()
	case len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h"):
		runHelp()
	case len(os.Args) > 1 && os.Args[1] == "--version":
		fmt.Println("rss-digest v1.0.0")
	default:
		runCLI()
	}
}
