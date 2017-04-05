package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	storage "google.golang.org/api/storage/v1"

	log "github.com/Sirupsen/logrus"
)

type (
	Storage interface {
		Download(bucket, object, destPath string) error
		Upload(bucket, object, srcPath string) error
	}

	CloudStorage struct {
		service *storage.ObjectsService
	}
)

func (ct *CloudStorage) Download(bucket, object, destPath string) error {
	logAttrs := log.Fields{"url": "gs://" + bucket + "/" + object, "destPath": destPath}
	log.WithFields(logAttrs).Debugln("Downloading")
	dest, err := os.Create(destPath)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Creating dest file")
		return err
	}
	defer dest.Close()

	resp, err := ct.service.Get(bucket, object).Download()
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to download")
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(dest, resp.Body)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to copy")
		return err
	}
	logAttrs["size"] = n
	log.WithFields(logAttrs).Debugln("Download successfully")
	return nil
}

func (ct *CloudStorage) Upload(bucket, object, srcPath string) error {
	logAttrs := log.Fields{"url": "gs://" + bucket + "/" + object, "srcPath": srcPath}
	log.WithFields(logAttrs).Debugln("Uploading")
	f, err := os.Open(srcPath)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to open the file")
		return err
	}
	_, err = ct.service.Insert(bucket, &storage.Object{Name: object}).Media(f).Do()
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to upload")
		return err
	}
	log.WithFields(logAttrs).Debugln("Upload successfully")
	return nil
}

type Target struct {
	Bucket    string
	Object    string
	LocalPath string
}

type TargetWorker struct {
	name    string
	targets chan *Target
	impl    func(bucket, object, srcPath string) error
	done    bool
	error   error
}

func (w *TargetWorker) run() {
	for {
		flds := log.Fields{}
		log.Debugln("Getting a target")
		var t *Target
		select {
		case t = <- w.targets:
		default: // Do nothing to break
		}
		if t == nil {
			log.Debugln("No target found any more")
			w.done = true
			w.error = nil
			break
		}

		flds["target"] = t
		log.WithFields(flds).Debugf("Start to %v\n", w.name)

		err := w.impl(t.Bucket, t.Object, t.LocalPath)
		flds["error"] = err
		if err != nil {
			log.WithFields(flds).Errorf("Failed to %v\n", w.name)
			w.done = true
			w.error = err
			break
		}
		log.WithFields(flds).Debugf("Finished to %v\n", w.name)
	}
}

type TargetWorkers []*TargetWorker

func (ws TargetWorkers) process(targets []*Target) error {
	c := make(chan *Target, len(targets))
	for _, t := range targets {
		c <- t
	}

	for _, w := range ws {
		w.targets = c
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

func (ws TargetWorkers) done() bool {
	for _, w := range ws {
		if !w.done {
			return false
		}
	}
	return true
}

func (ws TargetWorkers) error() error {
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
