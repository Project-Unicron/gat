package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(color.RedString("❌ Error:"), err)
		os.Exit(1)
	}
}
