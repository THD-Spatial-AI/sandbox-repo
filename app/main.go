package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/THD-Spatial/demo-repository/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	showV := flag.Bool("v", false, "print version and exit (shorthand)")
	flag.Parse()

	if *showV || *showVersion {
		fmt.Printf("%s (commit %s, built %s)\n", version.Version, version.Commit, version.Date)
		os.Exit(0)
	}

	fmt.Println("Hello world!")
}
