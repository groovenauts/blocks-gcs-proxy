package gcsproxy

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
	config, err := LoadConfig(configPath)
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

func LoadConfig(path string) (*ProcessConfig, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var res ProcessConfig
	err = json.Unmarshal(file, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
