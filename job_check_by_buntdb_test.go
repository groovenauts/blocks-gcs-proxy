package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type JobCheckCallee struct {
	Called bool
}

func (c *JobCheckCallee) Test() error {
	c.Called = true
	return nil
}

func TestJobCheckByBuntDB(t *testing.T) {
	c := &JobCheckByBuntDB{
		File:   "test-buntdb.db",
		Prefix: "jobs",
	}
	defer os.Remove(c.File)

	job_id1 := "job_id1"
	job_id2 := "job_id2"

	// 1st time
	(func() {
		ack := &JobCheckCallee{}
		main := &JobCheckCallee{}
		err := c.Check(job_id1, ack.Test, main.Test)
		assert.NoError(t, err)
		assert.Equal(t, false, ack.Called)
		assert.Equal(t, true, main.Called)
	})()

	// 2nd time
	(func() {
		ack := &JobCheckCallee{}
		main := &JobCheckCallee{}
		err := c.Check(job_id1, ack.Test, main.Test)
		assert.NoError(t, err)
		assert.Equal(t, true, ack.Called)
		assert.Equal(t, false, main.Called)
	})()

	// another job
	(func() {
		ack := &JobCheckCallee{}
		main := &JobCheckCallee{}
		err := c.Check(job_id2, ack.Test, main.Test)
		assert.NoError(t, err)
		assert.Equal(t, false, ack.Called)
		assert.Equal(t, true, main.Called)
	})()
}
