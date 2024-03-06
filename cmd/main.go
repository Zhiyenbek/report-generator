package main

import (
	"log"
	"os"
	"report-generator/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
