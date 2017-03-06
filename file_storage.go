package main

import (
	"io"
	"log"
	"os"

	storage "google.golang.org/api/storage/v1"
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
	log.Printf("Downloading gs://%v/%v to %v\n", bucket, object, destPath)
	dest, err := os.Create(destPath)
	if err != nil {
		log.Fatalf("Error creating %q: %v", destPath, err)
		return err
	}
	defer dest.Close()

	resp, err := ct.service.Get(bucket, object).Download()
	if err != nil {
		log.Printf("Error downloading bucket: %q object: %q because of %v", bucket, object, err)
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(dest, resp.Body)
	if err != nil {
		log.Fatalf("Error copry bucket: %q object: %q to %q because of %v", bucket, object, destPath, err)
		return err
	}
	log.Printf("Downloaded bucket: gs://%v/%v to %v (%d bytes)", bucket, object, destPath, n)
	return nil
}

func (ct *CloudStorage) Upload(bucket, object, srcPath string) error {
	log.Printf("Uploading %v to gs://%v/%v\n", srcPath, bucket, object)
	f, err := os.Open(srcPath)
	if err != nil {
		log.Fatalf("Error opening %q: %v", srcPath, err)
		return err
	}
	_, err = ct.service.Insert(bucket, &storage.Object{Name: object}).Media(f).Do()
	if err != nil {
		log.Printf("Error uploading gs://%q/%q: %v", bucket, srcPath, err)
		return err
	}
	log.Printf("Uploaded %v to gs://%v/%v\n", srcPath, bucket, object)
	return nil
}
