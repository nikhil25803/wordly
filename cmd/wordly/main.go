package main

import (
	"fmt"
	"log"
	"os"

	wordly "github.com/nikhil25803/wordly/internal"
	"github.com/nikhil25803/wordly/internal/db"
)

func main() {
	if err := db.ConnectDatabase(); err != nil {
		log.Fatal(err)
	}

	if err := wordly.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
