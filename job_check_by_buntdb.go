package main

import (
	"github.com/tidwall/buntdb"
)

type JobCheckByBuntDB struct {
	File   string
	Prefix string
}

func (jc *JobCheckByBuntDB) Check(job_id string, ack func() error, f func() error) error {
	key := jc.Prefix + job_id
	err := jc.Open(func(tx *buntdb.Tx) error {
		jobStatus, err := jc.GetStatus(tx, key)
		if err != nil {
			return err
		}
		if jobStatus != "" {
			log.Infof("Job %q is %s. So it will be skipped.\n", key, jobStatus)
			err := ack()
			if err != nil {
				log.Warningf("Failed to send ACK to skip Job %q.\n", key)
			}
			return nil
		}

		err = jc.SetStatus(tx, key, "executing", log.Errorf)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = f()

	return jc.Open(func(tx *buntdb.Tx) error {
		if err != nil {
			jc.SetStatus(tx, key, "error", log.Warningf)
			return err
		}

		jc.SetStatus(tx, key, "completed", log.Warningf)
		return nil
	})
}

func (jc *JobCheckByBuntDB) Open(f func(tx *buntdb.Tx) error) error {
	// Open the data.db file. It will be created if it doesn't exist.
	db, err := buntdb.Open(jc.File)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return db.Update(f)
}

func (jc *JobCheckByBuntDB) GetStatus(tx *buntdb.Tx, key string) (string, error) {
	val, err := tx.Get(key)
	if err == buntdb.ErrNotFound {
		return "", nil
	}
	if err != nil {
		log.Errorf("Failed to get value for %s because of %v\n", key, err)
		return "", err
	}
	return val, nil
}

func (jc *JobCheckByBuntDB) SetStatus(tx *buntdb.Tx, key, value string, logMethod func(string, ...interface{})) error {
	_, _, err := tx.Set(key, value, nil)
	if err != nil {
		logMethod("Failed to get value for %s because of %v\n", key, err)
		return err
	}
	return nil
}
