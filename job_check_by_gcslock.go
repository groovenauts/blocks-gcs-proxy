package main

import (
	"fmt"
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
	url := fmt.Sprintf("gs://%s/%s", jc.Bucket, object)

	logger := log.WithFields(logrus.Fields{"lock": url})
	logger.Infoln("JobCheckByGcslock Start")
	defer logger.Infoln("JobCheckByGcslock done")

	ok, err := jc.DeleteIfTimedout(object)
	if err != nil {
		log.Errorf("Failed to DeleteIfTimedout %s because of %v\n", url, err)
		return err
	}
	if !ok {
		log.Warningf("Quit running job because %s already exists\n", url)
		return nil
	}

	m, err := gcslock.New(nil, jc.Bucket, object)
	if err != nil {
		log.Errorf("Failed to gcslock.New because of %v\n", err)
		return err
	}

	if err := jc.Lock(m); err != nil {
		return err
	}
	defer jc.Unlock(m, func() error {
		err = jc.Storage.Delete(jc.Bucket, object)
		if err != nil {
			logger.Warningf("Deleting exceeded lock file instead of unlock error.\n")
			return err
		}
		logger.Infof("Deleting exceeded lock file instead of unlock success.\n")
		return nil
	})

	go jc.StartTouching(object, time.Duration(int64(jc.Timeout)/10))

	logger.Debugln("JobCheckByGcslock handler starting")
	defer logger.Debugln("JobCheckByGcslock handler done")

	err = f()

	jc.mux.Lock()
	defer jc.mux.Unlock()

	jc.working = false

	return err
}

func (jc *JobCheckByGcslock) DeleteIfTimedout(object string) (bool, error) {
	prefix := "JobCheckByGcslock.DeleteIfTimedout"
	logger := log.WithFields(logrus.Fields{"lock": fmt.Sprintf("gs://%s/%s", jc.Bucket, object)})
	logger.Debugf("%s Start\n", prefix)
	defer logger.Debugf("%s Done\n", prefix)

	f, err := jc.Storage.Get(jc.Bucket, object)
	if err != nil {
		return false, err
	}
	// Ok unless the file exists
	if f == nil {
		return true, nil
	}

	ut, err := time.Parse(time.RFC3339, f.Updated)
	if err != nil {
		logger.Errorf("%s error parsing file update time %q of %v because of %v\n", prefix, f.Updated, f, err)
		return false, err
	}

	deadline := ut.Add(jc.Timeout)
	if deadline.After(time.Now()) {
		// deadline hasn't come yet
		logger.Warningf("%s deadline hasn't come yet. It seems another process is working.\n", prefix)
		return false, nil
	}

	logger.Debugf("%s delete exceeded lock file starting.\n", prefix)
	err = jc.Storage.Delete(jc.Bucket, object)
	if err != nil {
		logger.Errorf("%s delete exceeded lock file error.\n", prefix)
		return false, err
	}
	logger.Infof("%s delete exceeded lock file successfully.\n", prefix)

	return true, nil
}

func (jc *JobCheckByGcslock) Lock(m gcslock.ContextLocker) error {
	log.Debugln("JobCheckByGcslock.Lock start")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait up to 10 seconds to acquire a lock.
	if err := m.ContextLock(ctx); err != nil {
		log.Errorf("Failed to ContextLock because of %v\n", err)
		return err
	}

	log.Infoln("JobCheckByGcslock.Lock done")
	return nil
}

func (jc *JobCheckByGcslock) Unlock(m gcslock.ContextLocker, f func() error) error {
	log.Debugln("JobCheckByGcslock.Unlock start")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.ContextUnlock(ctx); err != nil {
		if f == nil {
			log.Warningf("Failed to ContextUnlock because of %v\n", err)
			return err
		} else {
			return f()
		}
	}

	log.Infoln("JobCheckByGcslock.Unlock done")
	return nil
}

func (jc *JobCheckByGcslock) StartTouching(object string, interval time.Duration) error {
	logger := log.WithFields(logrus.Fields{"lock": fmt.Sprintf("gs://%s/%s", jc.Bucket, object), "interval": interval})
	logger.Infoln("JobCheckByGcslock.StartTouching Start")
	defer logger.Infoln("JobCheckByGcslock.StartTouching Finished")

	jc.working = true
	for {
		nextLimit := time.Now().Add(interval)
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

	logger := log.WithFields(logrus.Fields{"bucket": jc.Bucket, "object": object})
	logger.Debugln("WaitAndTouch")

	if !jc.working {
		logger.Infoln("WaitAndTouch working is done")
		return nil
	}

	metadata := map[string]string{
		"JobCheckByGcslock": time.Now().Format(time.RFC3339),
	}
	msg := &storage.Object{Metadata: metadata}

	logger = logger.WithFields(logrus.Fields{"metadata": metadata})
	logger.Debugln("WaitAndTouch Update lock file starting")
	_, err := jc.Storage.Update(jc.Bucket, object, msg)
	if err != nil {
		logger.Errorf("WaitAndTouch Update lock file error because of %v\n", err)
		return err
	}
	logger.Debugln("WaitAndTouch Update lock file success")
	return nil
}
