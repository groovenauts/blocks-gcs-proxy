package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	// log "github.com/Sirupsen/logrus"
)

type Workers []*Worker

func (ws Workers) process(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	lines := [][]byte{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Bytes())
	}

	c := make(chan []byte, len(lines))
	for _, line := range lines {
		c <- line
	}

	for _, w := range ws {
		w.lines = c
		go w.run()
	}

	for {
		time.Sleep(100 * time.Millisecond)
		if ws.done() {
			break
		}
	}

	return ws.error()
}

func (ws Workers) done() bool {
	for _, w := range ws {
		if !w.done {
			return false
		}
	}
	return true
}

func (ws Workers) error() error {
	messages := []string{}
	for _, w := range ws {
		if w.error != nil {
			messages = append(messages, w.error.Error())
		}
	}
	if len(messages) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(messages, "\n"))
}
