package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTargetWorker(t *testing.T) {
	type D1 struct {
		Error string
	}
	f := func(bucket, object, localPath string) error {
		log.Infof("f(%v, %v, %v)\n", bucket, object, localPath)
		if localPath == "" {
			log.Infof("localPath is blank\n")
			return fmt.Errorf("localPath is blank")
		}
		log.Infof("localPath is NOT blank\n")
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	workers := TargetWorkers{
		&TargetWorker{name: "w1", impl: f, maxTries: 1, interval: 1 * time.Second},
		&TargetWorker{name: "w2", impl: f, maxTries: 1, interval: 1 * time.Second},
		&TargetWorker{name: "w3", impl: f, maxTries: 1, interval: 1 * time.Second},
	}

	// Empty targets
	targets0 := Targets{}
	err := workers.process(targets0)
	assert.NoError(t, err)

	// 1 target
	targets1 := Targets{
		&Target{LocalPath: "foo"},
	}
	err = workers.process(targets1)
	assert.NoError(t, err)

	// 1 error target
	targets1Error := Targets{
		&Target{},
	}
	err = workers.process(targets1Error)
	assert.Equal(t, "localPath is blank", err.Error())

	// 1 success and 1 error
	targets1Success1Error := Targets{
		&Target{LocalPath: "foo"},
		&Target{},
	}
	err = workers.process(targets1Success1Error)
	assert.Equal(t, "localPath is blank", err.Error())

	// 3 success and 2 error
	targets3Success2Error := Targets{
		&Target{},
		&Target{LocalPath: "foo"},
		&Target{LocalPath: "foo"},
		&Target{},
		&Target{LocalPath: "foo"},
	}
	err = workers.process(targets3Success2Error)
	assert.Equal(t, "localPath is blank\nlocalPath is blank", err.Error())

	// 3 success
	targets3Success := Targets{
		&Target{LocalPath: "foo"},
		&Target{LocalPath: "foo"},
		&Target{LocalPath: "foo"},
	}
	err = workers.process(targets3Success)
	assert.NoError(t, err)
}
