package main

import (
	"strings"
)

type (
	RetryableError interface {
		Retryable() bool
	}
)

type (
	InvalidJobError struct {
		msg string
	}
)

func (e *InvalidJobError) Error() string {
	return e.msg
}

func (e *InvalidJobError) Retryable() bool {
	return false
}

type (
	CompositeError struct {
		errors []error
	}
)

func (e *CompositeError) Error() string {
	msgs := []string{}
	for _, err := range e.errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "\n")
}

func (e *CompositeError) Retryable() bool {
	for _, err := range e.errors {
		switch err.(type) {
		case RetryableError:
			e := err.(RetryableError)
			if !e.Retryable() {
				return false
			}
		}
	}
	return true
}

func (e *CompositeError) Any(f func(error) bool) bool {
	for _, err := range e.errors {
		if f(err) {
			return true
		}
	}
	return false
}
