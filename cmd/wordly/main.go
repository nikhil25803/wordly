package main

import (
	"fmt"
	"os"

	wordly "github.com/nikhil25803/wordly/internal"
)

func main() {
	if err := wordly.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
