package gcsproxy

import (
	"log"
	"os/exec"
	"strings"

	"golang.org/x/net/context"

	pubsub "google.golang.org/api/pubsub/v1"
)

type (
	JobConfig struct {
		Template []string
		Commands map[string][]string
		Dryrun bool
	}
	
	Job struct {
		config *JobConfig
		message *pubsub.ReceivedMessage
		notification *ProgressNotification
	}
)

func (job *Job) execute(ctx context.Context) error {
	cmd, err := job.build(ctx)
	if err != nil {
		log.Printf("Command build Error template: %v msg: %v cause of %v\n", job.config.Template, job.message, err)
		return err
	}
	err = cmd.Run()
	if err != nil {
		log.Printf("Command Error: cmd: %v cause of %v\n", cmd, err)
	}
	return nil
}

func (job *Job) build(ctx context.Context) (*exec.Cmd, error) {
	values, err := job.extract(ctx, job.config.Template)
	if err != nil {
		return nil, err
	}
	if len(job.config.Commands) > 0 {
		key := strings.Join(values, " ")
		t := job.config.Commands[key]
		if t == nil { t = job.config.Commands["default"] }
		if t != nil {
			values, err = job.extract(ctx, t)
			if err != nil {
				return nil, err
			}
		}
	}
	cmd := exec.Command(values[0], values[1:]...)
	return cmd, nil
}

func (job *Job) extract(ctx context.Context, values []string) ([]string, error) {
	result := []string{}
	for _, src := range values {
		extracted := src
		result = append(result, extracted)
	}
	return result, nil
}
