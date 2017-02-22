package gcsproxy

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	return job.setupWorkspace(ctx, func(workspace, downloads_dir, uploads_dir string) error {
		_, err := job.downloadFiles(downloads_dir)
		if err != nil {
			return err
		}

		cmd, err := job.build(ctx)
		if err != nil {
			log.Printf("Command build Error template: %v msg: %v cause of %v\n", job.config.Template, job.message, err)
			return err
		}
		err = cmd.Run()
		if err != nil {
			log.Printf("Command Error: cmd: %v cause of %v\n", cmd, err)
			// return err // Don't return this err
		}

		err = job.uploadFiles(uploads_dir)
		if err != nil {
			return err
		}

		return nil
	})
}

func (job *Job) setupWorkspace(ctx  context.Context, f func(workspace, downloads_dir, uploads_dir string) error) error {
	dir, err := ioutil.TempDir("", "workspace")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer os.RemoveAll(dir) // clean up

	subdirs := []string{
		filepath.Join(dir, "downloads"),
		filepath.Join(dir, "uploads"),
	}
	for _, subdir := range subdirs {
		err := os.MkdirAll(subdir, 0700)
		if err != nil {
			return err
		}
	}
	return f(dir, subdirs[0], subdirs[1])
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

func (job *Job) downloadFiles(dir string) (map[string]string, error) {
	result := map[string]string{}
	objects := job.flatten(job.parseJson(job.Message.Attributes["download_files"]))
	remote_files := []string{}
	for _, obj := range objects {
		switch obj.(type) {
		case string:
			remote_files = append(remote_files, obj.(string))
		default:
			log.Printf("Invalid download file URL: %v [%T]", obj, obj)
		}
	}

	return result, nil
}

func (job *Job) uploadFiles(dir string) error {
	return nil
}

func (job *Job) parseJson(source string) interface{} {
	matched, err := regexp.MatchString(`\A\[.*\]\z|\A\{.*\}\z|`, str)
	if err != nil {
		return str
	}
	if !matched {
		return str
	}
	var dest interface{}
	err = json.Unmarshal([]byte(str), &dest)
	if err != nil {
		return str
	}
	return dest
}


func (job *Job) flatten(obj interface{}) []interface{} {
	// Support only unmarshalled object from JSON
	// See https://golang.org/pkg/encoding/json/#Unmarshal also
	switch obj.(type) {
	case []interface{}:
		res := []interface{}{}
		for _, i := range obj {
			switch i.(type) {
			case bool, float64, string, nil:
				res = append(res, i)
			default:
				for _, j := range job.flatten(i) {
					res = append(res, j)
				}
			}
		}
		return res
	case map[string]interface{}:
		values := []interface{}
		for _, val := range obj {
			values = append(values, val)
		}
		return job.flatten(values)
	default:
		return []interface{}{obj}
	}
}
