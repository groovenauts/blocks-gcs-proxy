package main

import (
	"io"
	"os"

	storage "google.golang.org/api/storage/v1"

	logrus "github.com/sirupsen/logrus"
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
	logAttrs := logrus.Fields{"url": "gs://" + bucket + "/" + object, "destPath": destPath}
	log.WithFields(logAttrs).Debugln("Downloading")
	dest, err := os.Create(destPath)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnf("Creating dest file")
		return err
	}
	defer dest.Close()

	resp, err := ct.service.Get(bucket, object).Download()
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnf("Failed to download")
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(dest, resp.Body)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnf("Failed to copy")
		return err
	}
	logAttrs["size"] = n
	log.WithFields(logAttrs).Debugln("Download successfully")
	return nil
}

func (ct *CloudStorage) Upload(bucket, object, srcPath string) error {
	logAttrs := logrus.Fields{"url": "gs://" + bucket + "/" + object, "srcPath": srcPath}
	log.WithFields(logAttrs).Debugln("Uploading")
	f, err := os.Open(srcPath)
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnf("Failed to open the file")
		return err
	}
	_, err = ct.service.Insert(bucket, &storage.Object{Name: object}).Media(f).Do()
	if err != nil {
		logAttrs["error"] = err
		log.WithFields(logAttrs).Warnf("Failed to upload")
		return err
	}
	log.WithFields(logAttrs).Debugln("Upload successfully")
	return nil
}
