package main

import (
	"fmt"
	"os"

	"github.com/gusanmaz/onefile/cmd"
)

func main() {
	if err := cmd.NewPyPI2FileCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}