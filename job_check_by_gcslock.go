package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/context"

	storage "google.golang.org/api/storage/v1"

	"github.com/marcacohen/gcslock"
	logrus "github.com/sirupsen/logrus"
)

type JobCheckByGcslock struct {
	Bucket  string
	DirPath string
	Timeout time.Duration
	Storage Storage

	working bool
	mux     sync.Mutex
}

func (jc *JobCheckByGcslock) Check(job_id string, _ack func() error, f func() error) error {
	object := jc.DirPath + "/" + job_id + ".gcslock"

	ctx := context.Background()

	err := jc.DeleteIfTimedout(object)
	if err != nil {
		log.Errorf("Failed to DeleteIfTimedout gs://%s/%s because of %v\n", jc.Bucket, object, err)
		return err
	}

	m, err := gcslock.New(ctx, jc.Bucket, object)
	if err != nil {
		log.Errorf("Failed to gcslock.New because of %v\n", err)
		return err
	}

	if err := jc.Lock(ctx, m); err != nil {
		return err
	}
	defer jc.Unlock(ctx, m)

	go jc.StartTouching(object, time.Duration(int64(jc.Timeout)/10))

	err = f()

	jc.mux.Lock()
	defer jc.mux.Unlock()

	jc.working = false

	return err
}

func (jc *JobCheckByGcslock) DeleteIfTimedout(object string) error {
	f, err := jc.Storage.Get(jc.Bucket, object)
	if err != nil {
		if IsGoogleApiError(err, http.StatusNotFound) {
			return nil
		}
		return err
	}

	ut, err := time.Parse(time.RFC3339, f.Updated)
	if err != nil {
		log.Errorf("Failed to parse file update time %q of %v because of %v\n", f.Updated, f, err)
		return err
	}

	deadline := ut.Add(jc.Timeout)
	if deadline.After(time.Now()) {
		// deadline hasn't come yet
		return fmt.Errorf("Deadline hasn't come yet. It seems another process is working.")
	}

	err = jc.Storage.Delete(jc.Bucket, object)
	if err != nil {
		if !IsGoogleApiError(err, http.StatusNotFound) {
			return err
		}
	}

	return nil
}

func (jc *JobCheckByGcslock) Lock(c context.Context, m gcslock.ContextLocker) error {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	// Wait up to 1 second to acquire a lock.
	if err := m.ContextLock(ctx); err != nil {
		log.Errorf("Failed to ContextLock because of %v\n", err)
		return err
	}

	return nil
}

func (jc *JobCheckByGcslock) Unlock(c context.Context, m gcslock.ContextLocker) error {
	ctx, cancel := context.WithTimeout(c, 1*time.Second)
	defer cancel()

	if err := m.ContextUnlock(ctx); err != nil {
		log.Errorf("Failed to ContextUnlock because of %v\n", err)
		return err
	}
	return nil
}

func (jc *JobCheckByGcslock) StartTouching(object string, interval time.Duration) error {
	jc.working = true
	for {
		nextLimit := time.Now().Add(time.Duration(interval) * time.Second)
		err := jc.WaitAndTouch(object, nextLimit)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err}).Errorln("Error in StartTouching")
			return err
		}
		if !jc.working {
			return nil
		}
	}
	// return nil
}

func (jc *JobCheckByGcslock) WaitAndTouch(object string, nextLimit time.Time) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	for now := range ticker.C {
		if !jc.working {
			ticker.Stop()
			return nil
		}
		if now.After(nextLimit) {
			ticker.Stop()
			break
		}
	}

	jc.mux.Lock()
	defer jc.mux.Unlock()

	logAttrs := logrus.Fields{"bucket": jc.Bucket, "object": object}
	log.WithFields(logAttrs).Debugln("WaitAndTouch")

	if !jc.working {
		log.WithFields(logAttrs).Infoln("WaitAndTouch working is done")
		return nil
	}

	metadata := &storage.Object{
		Metadata: map[string]string{
			"JobCheckByGcslock": time.Now().Format(time.RFC3339),
		},
	}
	_, err := jc.Storage.Update(jc.Bucket, object, metadata)
	if err != nil {
		if IsGoogleApiError(err, http.StatusNotFound) {
			return nil
		}
		return err
	}
	return nil
}
