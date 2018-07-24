package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlParse(t *testing.T) {
	(func() {
		s := "gs://bucket0/path/to/file0"
		r, err := url.Parse(s)
		assert.NoError(t, err)
		assert.Equal(t, "gs", r.Scheme)
		assert.Equal(t, "bucket0", r.Host)
		assert.Equal(t, "/path/to/file0", r.Path)

		r2, err := urlParse(s)
		assert.NoError(t, err)
		assert.Equal(t, "gs", r2.Scheme)
		assert.Equal(t, "bucket0", r2.Host)
		assert.Equal(t, "/path/to/file0", r2.Path)

		assert.Equal(t, r, r2)
	})()

	(func() {
		s := "gs://bucket0/path%/to/file0"
		_, err := url.Parse(s)
		assert.Error(t, err)

		r2, err := urlParse(s)
		assert.NoError(t, err)
		assert.Equal(t, "gs", r2.Scheme)
		assert.Equal(t, "bucket0", r2.Host)
		assert.Equal(t, "/path%/to/file0", r2.Path)
	})()
}
