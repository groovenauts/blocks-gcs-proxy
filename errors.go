package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type (
	RetryableError interface {
		Retryable() bool
	}

	NestableError interface {
		CausedBy(err error) bool
	}
)

type (
	InvalidJobError struct {
		msg   string
		cause error
	}
)

func SameErrorType(obj, expected error) bool {
	et1 := reflect.TypeOf(obj)
	et2 := reflect.TypeOf(expected)
	if et1 == et2 {
		return true
	}
	switch obj.(type) {
	case NestableError:
		ne := obj.(NestableError)
		return ne.CausedBy(expected)
	default:
		return false
	}
}

func (e *InvalidJobError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	if e.cause != nil {
		return e.cause.Error()
	}
	return ""
}

func (e *InvalidJobError) Retryable() bool {
	return false
}

func (e *InvalidJobError) CausedBy(err error) bool {
	return SameErrorType(e.cause, err)
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

func (e *CompositeError) CausedBy(err error) bool {
	f := func(e error) bool {
		return SameErrorType(e, err)
	}
	return e.Any(f)
}

func (e *CompositeError) Any(f func(error) bool) bool {
	for _, err := range e.errors {
		if f(err) {
			return true
		}
	}
	return false
}

type ConfigError struct {
	Name      string
	Ancestors []string
	Message   string
}

func (e *ConfigError) Setup() {
	if e.Ancestors == nil {
		e.Ancestors = []string{e.Name}
	}
}

func (e *ConfigError) Add(ancestor string) {
	e.Setup()
	e.Ancestors = append(e.Ancestors, ancestor)
}

func (e *ConfigError) Error() string {
	e.Setup()
	sort.Sort(sort.Reverse(sort.StringSlice(e.Ancestors)))
	return fmt.Sprintf("%s %s", strings.Join(e.Ancestors, "."), e.Message)
}

type ConfigSetup func()*ConfigError
