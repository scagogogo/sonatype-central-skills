package main

import (
	"fmt"
	"os"

	"github.com/scagogogo/sonatype-central-sdk/cmd/sonatype-central/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}