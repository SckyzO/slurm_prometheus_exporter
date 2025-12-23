package main

import (
	"fmt"
	"os"
)

// Version is the version of the exporter, set during build time via -ldflags
var Version = "dev"

func main() {
	fmt.Printf("Slurm Exporter v%s\n", Version)
	os.Exit(0)
}
