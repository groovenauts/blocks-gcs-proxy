package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

type TestAppInvocation struct {
	Cmd  string
	Args []string
}

func NewAppTestRun(args []string) []*TestAppInvocation {
	app := newApp()
	return AppTestRun(app, args)
}

func AppTestRun(app *cli.App, args []string) []*TestAppInvocation {
	invocations := []*TestAppInvocation{}

	newCommands := []cli.Command{}
	for _, cmd := range app.Commands {
		newCmd := cmd
		newCmd.Action = func(c *cli.Context) error {
			invocations = append(invocations, &TestAppInvocation{cmd.Name, []string(c.Args())})
			return nil
		}
		newCommands = append(newCommands, newCmd)
	}
	app.Commands = newCommands

	app.Action = func(c *cli.Context) error {
		invocations = append(invocations, &TestAppInvocation{"main", []string(c.Args())})
		return nil
	}
	app.Run(args)
	return invocations
}

func TestNewAppMain01(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "./app.sh", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "main", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "foo", "bar"}, i.Args)
}

func TestNewAppMain02(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "-c", "config.json", "./app.sh", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "main", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "foo", "bar"}, i.Args)
}

func TestNewAppMain03(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "./app.sh", "-c", "config.json", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "main", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "-c", "config.json", "foo", "bar"}, i.Args)
}

func TestNewAppMain04(t *testing.T) {
	app := newApp()
	app.Writer = ioutil.Discard
	invocations := AppTestRun(app, []string{"./blocks-gcs-proxy", "-X", "XXX", "./app.sh", "foo", "bar"})
	assert.Equal(t, 0, len(invocations))
	// cli shows "Incorrect Usage. flag provided but not defined: -X"
}

func TestNewAppMain05(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "./app.sh", "-X", "XXX", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "main", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "-X", "XXX", "foo", "bar"}, i.Args)
}

func TestNewAppExec01(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "exec", "./app.sh", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "exec", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "foo", "bar"}, i.Args)
}

func TestNewAppExec02(t *testing.T) {
	invocations := NewAppTestRun([]string{"./blocks-gcs-proxy", "exec", "-c", "config.json", "./app.sh", "foo", "bar"})
	assert.Equal(t, 1, len(invocations))
	i := invocations[0]
	assert.Equal(t, "exec", i.Cmd)
	assert.Equal(t, []string{"./app.sh", "foo", "bar"}, i.Args)
}
