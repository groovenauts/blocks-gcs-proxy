package gcsproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"net/url"

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
		storage Storage
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
	objects := job.flatten(job.parseJson(job.message.Message.Attributes["download_files"]))
	remote_urls := []string{}
	for _, obj := range objects {
		switch obj.(type) {
		case string:
			remote_urls = append(remote_urls, obj.(string))
		default:
			log.Printf("Invalid download file URL: %v [%T]", obj, obj)
		}
	}
	for _, remote_url := range remote_urls {
		url, err := url.Parse(remote_url)
		if err != nil {
			log.Fatalf("Invalid URL: %v because of %v\n", remote_url, err)
			return nil, err
		}
		urlstr := fmt.Sprintf("gs://%v/%v", url.Host, url.Path)
		destPath := filepath.Join(dir, url.Host, url.Path)
		err = job.storage.Download(url.Host, url.Path, destPath)
		if err != nil {
			return nil, err
		}
		result[urlstr] = destPath
	}
	return result, nil
}

func (job *Job) uploadFiles(dir string) error {
	localPaths, err := job.listFiles(dir)
	if err != nil {
		return err
	}
	for _, localPath := range localPaths {
		relPath, err := filepath.Rel(dir, localPath)
		if err != nil {
			log.Fatalf("Error getting relative path of %v: %v\n", localPath, err)
			return err
		}
		sep := string([]rune{os.PathSeparator})
		parts := strings.Split(relPath, sep)
		bucket := parts[0]
		object := strings.Join(parts[1:], sep)
		err = job.storage.Upload(bucket, object, localPath)
		if err != nil {
			log.Fatalf("Error uploading %v to gs://%v/%v: %v\n", localPath, bucket, object, err)
			return err
		}
	}
	return nil
}

func (job *Job) listFiles(dir string) ([]string, error) {
	result := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error listing upload files: %v\n", err)
		return nil, err
	}
	return result, nil
}

func (job *Job) parseJson(str string) interface{} {
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
		for _, i := range obj.([]interface{}) {
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
		values := []interface{}{}
		for _, val := range obj.(map[string]interface{}) {
			values = append(values, val)
		}
		return job.flatten(values)
	default:
		return []interface{}{obj}
	}
}
