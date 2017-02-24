package gcsproxy

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
	dest, err := os.Create(destPath)
	if err != nil {
		log.Fatalf("Error creating %q: %v", destPath, err)
		return err
	}
	defer dest.Close()

	resp, err := ct.service.Get(bucket, object).Download()
	if err != nil {
		log.Printf("Error downloading gs://%q/%q: %v", bucket, object, err)
		return err
	}
	defer resp.Body.Close()

	n, err := io.Copy(dest, resp.Body)
	if err != nil {
		log.Fatalf("Error copry gs://%q/%q to %q: %v", bucket, object, destPath, err)
		return err
	}
	log.Printf("Downloaded gs://%q/%q: %d bytes", bucket, object, n)
	return nil
}

func (ct *CloudStorage) Upload(bucket, object, srcPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		log.Fatalf("Error opening %q: %v", srcPath, err)
		return err
	}
	obj, err := ct.service.Insert(bucket, &storage.Object{Name: object}).Media(f).Do()
	log.Printf("Got storage.Object, err: %#v, %v", obj, err)
	if err != nil {
		log.Printf("Error uploading gs://%q/%q: %v", bucket, srcPath, err)
		return err
	}
	return nil
}
