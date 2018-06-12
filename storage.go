package main

import (
	"io"
	"mime"
	"os"
	"path"

	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"

	logrus "github.com/sirupsen/logrus"
)

type (
	Storage interface {
		Download(bucket, object, destPath string) error
		Upload(bucket, object, srcPath string) error
		Get(bucket, object string) (*storage.Object, error)
		Delete(bucket, object string) error
		Update(bucket, object string, body *storage.Object) (*storage.Object, error)
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

func (ct *CloudStorage) Get(bucket, object string) (*storage.Object, error) {
	log := log.WithFields(logrus.Fields{"url": "gs://" + bucket + "/" + object})
	log.Debugln("Getting file info")
	obj, err := ct.service.Get(bucket, object).Do()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Errorf("Failed to get GCS file info")
		return nil, err
	}
	return obj, err
}

func (ct *CloudStorage) Delete(bucket, object string) error {
	log := log.WithFields(logrus.Fields{"url": "gs://" + bucket + "/" + object})
	log.Debugln("Deleting file")
	err := ct.service.Delete(bucket, object).Do()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Errorf("Failed to delete GCS file")
		return err
	}
	return err
}

func (ct *CloudStorage) Update(bucket, object string, body *storage.Object) (*storage.Object, error) {
	log := log.WithFields(logrus.Fields{"url": "gs://" + bucket + "/" + object})
	log.Debugln("Updating file")
	obj, err := ct.service.Update(bucket, object, body).Do()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Errorf("Failed to update GCS file")
		return nil, err
	}
	return obj, nil
}

func IsGoogleApiError(err error, code int) bool {
	if err != nil {
		apiErr, ok := err.(*googleapi.Error)
		if ok {
			if apiErr.Code == code {
				return true
			}
		}
	}
	return false
}
