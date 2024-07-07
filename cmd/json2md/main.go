package main

import (
	"fmt"
	"os"

	"github.com/gusanmaz/onefile/cmd"
)

func main() {
	if err := cmd.NewJSON2MDCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
