package main

import (
	"io"
	"mime"
	"os"
	"path"

	storage "google.golang.org/api/storage/v1"

	logrus "github.com/sirupsen/logrus"
)

type (
	Storage interface {
		Download(bucket, object, destPath string) error
		Upload(bucket, object, srcPath string) error
	}

	CloudStorage struct {
		service          *storage.ObjectsService
		ContentTypeByExt bool
	}
)

func (ct *CloudStorage) Download(bucket, object, destPath string) error {
	log := log.WithFields(logrus.Fields{"url": "gs://" + bucket + "/" + object, "destPath": destPath})
	log.Debugln("Downloading")
	dest, err := os.Create(destPath)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Warnf("Creating dest file")
		return err
	}
	defer dest.Close()

	resp, err := ct.service.Get(bucket, object).Download()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Warnf("Failed to download")
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(dest, resp.Body)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Warnf("Failed to copy")
		return err
	}
	log.WithFields(logrus.Fields{"size": n}).Debugln("Download successfully")
	return nil
}

func (ct *CloudStorage) Upload(bucket, object, srcPath string) error {
	logAttrs := logrus.Fields{"url": "gs://" + bucket + "/" + object, "srcPath": srcPath}
	log.WithFields(logAttrs).Debugln("Uploading")
	f, err := os.Open(srcPath)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Warnf("Failed to open the file")
		return err
	}
	obj := &storage.Object{Name: object}
	if ct.ContentTypeByExt {
		obj.ContentType = mime.TypeByExtension(path.Ext(object))
	}
	_, err = ct.service.Insert(bucket, obj).Media(f).Do()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Warnf("Failed to upload")
		return err
	}
	log.WithFields(logAttrs).Debugln("Upload successfully")
	return nil
}
