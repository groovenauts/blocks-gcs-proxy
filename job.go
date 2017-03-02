package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
)

type (
	CommandConfig struct {
		Template []string            `json:"-"`
		Options  map[string][]string `json:"options,omitempty"`
		Dryrun   bool                `json:"dryrun,omitempty"`
	}

	Job struct {
		config *CommandConfig
		// https://godoc.org/google.golang.org/genproto/googleapis/pubsub/v1#ReceivedMessage
		message      *JobMessage
		notification *ProgressNotification
		storage      Storage

		// These are set at at setupWorkspace
		workspace     string
		downloads_dir string
		uploads_dir   string

		// These are set at setupDownloadFiles
		downloadFileMap     map[string]string
		remoteDownloadFiles interface{}
		localDownloadFiles  interface{}

		cmd *exec.Cmd
	}
)

func (job *Job) run(ctx context.Context) error {
	err := job.runImpl(ctx)
	switch err.(type) {
	case InvalidJobError:
		err := job.withNotify(CANCELLING, job.message.Ack)()
		if err != nil {
			return err
		}
	}
	return err
}

func (job *Job) runImpl(ctx context.Context) error {
	job.notification.notify(PROCESSING, job.message.MessageId(), "info")
	defer job.withNotify(CLEANUP, job.clearWorkspace)() // Call clearWorkspace even if setupWorkspace retuns error

	err := job.message.Validate()
	if err != nil {
		log.Printf("Invalid Message: MessageId: %v, Message: %v, error: %v\n", job.message.MessageId(), job.message.raw.Message, err)
		return nil
	}

	err = job.setupWorkspace()
	if err != nil {
		return err
	}

	err = job.withNotify(PREPARING, job.setupDownloadFiles)()
	if err != nil {
		return err
	}

	err = job.build()
	if err != nil {
		log.Fatalf("Command build Error template: %v msg: %v cause of %v\n", job.config.Template, job.message, err)
		return err
	}

	go job.message.sendMADPeriodically()
	defer job.message.Done()

	err = job.withNotify(DOWNLOADING, job.downloadFiles)()
	if err != nil {
		return err
	}

	err = job.withNotify(EXECUTING, job.execute)()
	if err != nil {
		return err
	}

	err = job.withNotify(UPLOADING, job.uploadFiles)()
	if err != nil {
		return err
	}

	err = job.withNotify(ACKSENDING, job.message.Ack)()
	if err != nil {
		return err
	}

	return err
}

func (job *Job) withNotify(progress int, f func() error) func() error {
	msg_id := job.message.MessageId()
	return func() error {
		job.notification.notify(progress, msg_id, "info")
		err := f()
		if err != nil {
			job.notification.notify(progress+2, msg_id, "error")
			return err
		}
		job.notification.notify(progress+1, msg_id, "info")
		return nil
	}
}

func (job *Job) setupWorkspace() error {
	dir, err := ioutil.TempDir("", "workspace")
	if err != nil {
		log.Fatal(err)
		return err
	}

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
	job.workspace = dir
	job.downloads_dir = subdirs[0]
	job.uploads_dir = subdirs[1]
	return nil
}

func (job *Job) clearWorkspace() error {
	if job.workspace != "" {
		return os.RemoveAll(job.workspace)
	}
	return nil
}

func (job *Job) setupDownloadFiles() error {
	job.downloadFileMap = map[string]string{}
	job.remoteDownloadFiles = job.message.DownloadFiles()
	objects := job.flatten(job.remoteDownloadFiles)
	remoteUrls := []string{}
	for _, obj := range objects {
		switch obj.(type) {
		case string:
			remoteUrls = append(remoteUrls, obj.(string))
		default:
			log.Printf("Invalid download file URL: %v [%T]", obj, obj)
		}
	}
	for _, remote_url := range remoteUrls {
		url, err := url.Parse(remote_url)
		if err != nil {
			log.Fatalf("Invalid URL: %v because of %v\n", remote_url, err)
			return err
		}
		urlstr := fmt.Sprintf("gs://%v%v", url.Host, url.Path)
		destPath := filepath.Join(job.downloads_dir, url.Host, url.Path)
		job.downloadFileMap[urlstr] = destPath
	}
	job.localDownloadFiles = job.copyWithFileMap(job.remoteDownloadFiles)
	return nil
}

func (job *Job) copyWithFileMap(obj interface{}) interface{} {
	switch obj.(type) {
	case map[string]interface{}:
		result := map[string]interface{}{}
		for k, v := range obj.(map[string]interface{}) {
			result[k] = job.copyWithFileMap(v)
		}
		return result
	case []interface{}:
		result := []interface{}{}
		for _, v := range obj.([]interface{}) {
			result = append(result, job.copyWithFileMap(v))
		}
		return result
	case string:
		return job.downloadFileMap[obj.(string)]
	default:
		return obj
	}
}

func (job *Job) buildVariable() *Variable {
	return &Variable{
		data: map[string]interface{}{
			"workspace":             job.workspace,
			"downloads_dir":         job.downloads_dir,
			"uploads_dir":           job.uploads_dir,
			"download_files":        job.localDownloadFiles,
			"local_download_files":  job.localDownloadFiles,
			"remote_download_files": job.remoteDownloadFiles,
			"attrs":                 job.message.raw.Message.Attributes,
			"attributes":            job.message.raw.Message.Attributes,
			"data":                  job.message.raw.Message.Data,
		},
	}
}

func (job *Job) build() error {
	v := job.buildVariable()

	values, err := job.extract(v, job.config.Template)
	if err != nil {
		return err
	}
	if len(job.config.Options) > 0 {
		key := strings.Join(values, " ")
		t := job.config.Options[key]
		if t == nil {
			t = job.config.Options["default"]
		}
		if t != nil {
			values, err = job.extract(v, t)
			if err != nil {
				return err
			}
		}
	}
	job.cmd = exec.Command(values[0], values[1:]...)
	return nil
}

func (job *Job) extract(v *Variable, values []string) ([]string, error) {
	result := []string{}
	for _, src := range values {
		extracted, err := v.expand(src)
		if err != nil {
			return nil, &InvalidJobError{err.Error()}
		}
		vals := strings.Split(extracted, v.separator)
		for _, val := range vals {
			result = append(result, val)
		}
	}
	return result, nil
}

func (job *Job) downloadFiles() error {
	for remoteURL, destPath := range job.downloadFileMap {
		url, err := url.Parse(remoteURL)
		if err != nil {
			log.Fatalf("Invalid URL: %v because of %v\n", remoteURL, err)
			return err
		}

		dir := path.Dir(destPath)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}

		err = job.storage.Download(url.Host, url.Path[1:], destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (job *Job) execute() error {
	var out bytes.Buffer
	job.cmd.Stdout = &out
	job.cmd.Stderr = &out
	log.Printf("EXECUTE running: %v\n", job.cmd)
	err := job.cmd.Run()
	if err != nil {
		log.Printf("Command Error: cmd: %v cause of %v\n%v\n", job.cmd, err, out.String())
		return err
	}
	return nil
}

func (job *Job) uploadFiles() error {
	localPaths, err := job.listFiles(job.uploads_dir)
	if err != nil {
		return err
	}
	for _, localPath := range localPaths {
		relPath, err := filepath.Rel(job.uploads_dir, localPath)
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
