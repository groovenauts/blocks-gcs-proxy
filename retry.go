package main

import (
	"time"

	"github.com/cenkalti/backoff"
)

func RetryWithSleep(operation backoff.Operation, b backoff.BackOff) error {
	var err error
	var next time.Duration

	b.Reset()
	for {
		if err = operation(); err == nil {
			return nil
		}

		if permanent, ok := err.(*backoff.PermanentError); ok {
			return permanent.Err
		}

		if next = b.NextBackOff(); next == backoff.Stop {
			return err
		}

		time.Sleep(next)
	}
}
