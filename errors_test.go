package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	TestError1 struct {
		msg string
	}
)

func (e *TestError1) Error() string {
	return e.msg
}

type (
	TestError2 struct {
		msg string
		cause error
	}
)

func (e *TestError2) Error() string {
	return e.msg
}

func (e *TestError2) CausedBy(err error) bool {
	t0 := reflect.TypeOf(e.cause)
	t1 := reflect.TypeOf(err)
	return t0 == t1
}

func TestCausedBy(t *testing.T) {
	e1 := fmt.Errorf("Something wrong")
	e2 := &TestError1{"Test Error"}
	e3 := &TestError2{cause: e2}
	err0 := &InvalidJobError{}
	err1 := &InvalidJobError{cause: e1}
	err2 := &InvalidJobError{cause: e2}
	err3 := &InvalidJobError{cause: e3}

	assert.False(t, err0.CausedBy((*TestError1)(nil)))
	assert.False(t, err1.CausedBy((*TestError1)(nil)))
	assert.True(t, err2.CausedBy((*TestError1)(nil)))
	assert.True(t, err3.CausedBy((*TestError1)(nil)))

	assert.False(t, (&CompositeError{[]error{}}).CausedBy((*TestError1)(nil)))
	assert.False(t, (&CompositeError{[]error{e1}}).CausedBy((*TestError1)(nil)))
	assert.True(t, (&CompositeError{[]error{e2}}).CausedBy((*TestError1)(nil)))
	assert.True(t, (&CompositeError{[]error{e3}}).CausedBy((*TestError1)(nil)))

	assert.False(t, (&CompositeError{[]error{err0}}).CausedBy((*TestError1)(nil)))
	assert.False(t, (&CompositeError{[]error{err1}}).CausedBy((*TestError1)(nil)))
	assert.True(t, (&CompositeError{[]error{err2}}).CausedBy((*TestError1)(nil)))
	assert.True(t, (&CompositeError{[]error{err3}}).CausedBy((*TestError1)(nil)))
}
