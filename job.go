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

		// These are set at at setupWorkspace
		workspace string
		downloads_dir string
		uploads_dir string

		// These are set at setupDownloadFiles
		downloadFileMap map[string]string
		remoteDownloadFiles interface{}
		localDownloadFiles interface{}
	}
)


func (job *Job) execute(ctx context.Context) error {
	return job.setupWorkspace(ctx, func() error {
		err := job.setupDownloadFiles()
		if err != nil { return err }

		err = job.downloadFiles()
		if err != nil { return err }

		cmd, err := job.build()
		if err != nil {
			log.Printf("Command build Error template: %v msg: %v cause of %v\n", job.config.Template, job.message, err)
			return err
		}
		err = cmd.Run()
		if err != nil {
			log.Printf("Command Error: cmd: %v cause of %v\n", cmd, err)
			// return err // Don't return this err
		}

		err = job.uploadFiles()
		if err != nil { return err }

		return nil
	})
}


func (job *Job) setupWorkspace(ctx  context.Context, f func() error) error {
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
	job.workspace = dir
	job.downloads_dir = subdirs[0]
	job.uploads_dir = subdirs[1]
	return f()
}


func (job *Job) setupDownloadFiles() error {
	job.downloadFileMap = map[string]string{}
	job.remoteDownloadFiles = job.parseJson(job.message.Message.Attributes["download_files"])
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
		urlstr := fmt.Sprintf("gs://%v/%v", url.Host, url.Path)
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
	default:
		return obj
	}
}

func (job *Job) downloadFiles() error {
	for remoteURL, destPath := range job.downloadFileMap {
		url, err := url.Parse(remoteURL)
		if err != nil {
			log.Fatalf("Invalid URL: %v because of %v\n", remoteURL, err)
			return err
		}
		err = job.storage.Download(url.Host, url.Path, destPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (job *Job) build() (*exec.Cmd, error) {
	values, err := job.extract(job.config.Template)
	if err != nil {
		return nil, err
	}
	if len(job.config.Commands) > 0 {
		key := strings.Join(values, " ")
		t := job.config.Commands[key]
		if t == nil { t = job.config.Commands["default"] }
		if t != nil {
			values, err = job.extract(t)
			if err != nil {
				return nil, err
			}
		}
	}
	cmd := exec.Command(values[0], values[1:]...)
	return cmd, nil
}

func (job *Job) extract(values []string) ([]string, error) {
	result := []string{}
	v := Variable{
		data: map[string]interface{}{
			"workspace": job.workspace,
			"downloads_dir": job.downloads_dir,
			"uploads_dir": job.uploads_dir,
			"download_files": job.localDownloadFiles,
			"local_download_files": job.localDownloadFiles,
			"remote_download_files": job.remoteDownloadFiles,
			"attrs": job.message.Message.Attributes,
			"attributes": job.message.Message.Attributes,
			"data": job.message.Message.Data,
		},
	}
	for _, src := range values {
		extracted, err := v.expand(src)
		if err != nil {return nil, err}
		result = append(result, extracted)
	}
	return result, nil
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
