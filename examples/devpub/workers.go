package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Workers []*Worker

func (ws Workers) process(filepath string) error {
	flds := log.Fields{"filepath": filepath}
	log.WithFields(flds).Debugln("Workers processing")

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	log.WithFields(flds).Debugln("Workers scanning")
	lines := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	flds["line_counts"] = len(lines)
	log.WithFields(flds).Debugln("Workers sending lines to channel")
	c := make(chan string, len(lines))
	for _, line := range lines {
		c <- line
	}

	log.WithFields(flds).Debugln("Workers starting each worker")
	for _, w := range ws {
		w.lines = c
		go w.run()
	}

	log.WithFields(flds).Debugln("Workers waiting")
	for {
		time.Sleep(100 * time.Millisecond)
		if ws.done() {
			break
		}
	}

	log.WithFields(flds).Debugln("Workers finishing")
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
