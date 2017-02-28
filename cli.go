package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/net/context"
)

func main() {
	ctx := context.Background()

	configPath := "./config.json"
	config, err := LoadProcessConfig(configPath)
	if err != nil {
		fmt.Printf("Error to load %v cause of %v\n", configPath, err)
		os.Exit(1)
	}
	config.setup(ctx, os.Args)

	p := &Process{config: config}
	err = p.setup(ctx)
	if err != nil {
		fmt.Printf("Error to setup Process cause of %v\n", err)
		os.Exit(1)
	}

	err = p.run(ctx)
	if err != nil {
		fmt.Printf("Error to run cause of %v\n", err)
		os.Exit(1)
	}
}
