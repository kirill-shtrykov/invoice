package main

import (
	"fmt"
	"log"
	"os"
)

func run() error {
	config, err := LoadConfig(defaultConfigFile)
	if err != nil {
		return fmt.Errorf("unable load config: %v", err)
	}
	return CreatePDF(config)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
