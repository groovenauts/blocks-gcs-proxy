package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/groovenauts/blocks-variable"

	log "github.com/Sirupsen/logrus"
)

type CommandConfig struct {
	Template    []string            `json:"-"`
	Options     map[string][]string `json:"options,omitempty"`
	Dryrun      bool                `json:"dryrun,omitempty"`
	Uploaders   int                 `json:"uploaders,omitempty"`
	Downloaders int                 `json:"downloaders,omitempty"`
}

func (c *CommandConfig) setup() {
	if c.Downloaders < 1 {
		c.Downloaders = 1
	}
	if c.Uploaders < 1 {
		c.Uploaders = 1
	}
}

type Job struct {
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

func (job *Job) run() error {
	err := job.runWithoutErrorHandling()
	switch err.(type) {
	case RetryableError:
		var f func() error
		e := err.(RetryableError)
		if e.Retryable() {
			f = job.withNotify(NACKSENDING, job.message.Nack)
		} else {
			f = job.withNotify(CANCELLING, job.message.Ack)
		}
		err := f()
		if err != nil {
			return err
		}
	}
	return err
}

func (job *Job) runWithoutErrorHandling() error {
	defer job.withNotify(CLEANUP, job.clearWorkspace)() // Call clearWorkspace even if setupWorkspace retuns error

	err := job.withNotify(INITIALIZING, job.prepare)()
	if err != nil {
		return err
	}

	go job.message.sendMADPeriodically(job.notification)
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

func (job *Job) withNotify(step JobStep, f func() error) func() error {
	return job.notification.wrap(job.message.MessageId(), step, f)
}

func (job *Job) prepare() error {
	logAttrs := log.Fields{"job_message_id": job.message.MessageId()}
	err := job.message.Validate()
	if err != nil {
		logAttrs["message"] = job.message.raw.Message
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Invalid Message")
		return err
	}

	err = job.setupWorkspace()
	if err != nil {
		return err
	}

	err = job.useDataAsAttributesIfPossible()
	if err != nil {
		return err
	}

	job.remoteDownloadFiles = job.message.DownloadFiles()
	err = job.setupDownloadFiles()
	if err != nil {
		return err
	}

	err = job.build()
	if err != nil {
		logAttrs["template"] = job.config.Template
		logAttrs["message"] = job.message
		logAttrs["error"] = err
		log.WithFields(logAttrs).Errorf("Failed to build command")
		return err
	}
	return nil
}

func (job *Job) setupWorkspace() error {
	if job.workspace != "" {
		return nil
	}
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

const UseDataAsAttributesKey = "use-data-as-attributes"

var UseDataAsAttributesRegexp = regexp.MustCompile("(?i)true|yes|on|1")

func (job *Job) useDataAsAttributesIfPossible() error {
	attrs := job.message.raw.Message.Attributes
	flg := attrs[UseDataAsAttributesKey]
	if !UseDataAsAttributesRegexp.MatchString(flg) {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(job.message.raw.Message.Data)
	if err != nil {
		logAttrs := log.Fields{"error": err, "data": job.message.raw.Message.Data}
		log.WithFields(logAttrs).Errorf("Failed to decode by base64")
		return err
	}
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(decoded), &parsed)
	if err != nil {
		logAttrs := log.Fields{"error": err, "data": job.message.raw.Message.Data}
		log.WithFields(logAttrs).Errorf("Failed to json.Unmarshal")
		return err
	}
	for key, obj := range parsed {
		var value string
		switch obj.(type) {
		case string:
			value = obj.(string)
		default:
			b, err := json.Marshal(obj)
			if err != nil {
				logAttrs := log.Fields{"error": err, "obj": obj}
				log.WithFields(logAttrs).Errorf("Failed to json.Marshal")
				return err
			}
			value = string(b)
		}
		attrs[key] = value
	}
	return nil
}

func (job *Job) setupDownloadFiles() error {
	job.downloadFileMap = map[string]string{}
	objects := job.flatten(job.remoteDownloadFiles)
	remoteUrls := []string{}
	for _, obj := range objects {
		switch obj.(type) {
		case string:
			remoteUrls = append(remoteUrls, obj.(string))
		default:
			log.WithFields(log.Fields{"url": obj}).Errorf("Invalid download file URL: %T\n", obj)
		}
	}
	for _, remote_url := range remoteUrls {
		url, err := url.Parse(remote_url)
		if err != nil {
			log.WithFields(log.Fields{"url": remote_url}).Errorln("Invalid download file URL")
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

func (job *Job) buildVariable() *bvariable.Variable {
	return &bvariable.Variable{
		Data: map[string]interface{}{
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
	logAttrs := log.Fields{}
	v := job.buildVariable()
	values, err := job.extract(v, job.config.Template)
	if len(job.config.Options) > 0 {
		logAttrs["options_key_template"] = job.config.Template
		logAttrs["options_key_base"] = values
		log.WithFields(logAttrs).Debugln("extracting key of options")
		if err != nil {
			logAttrs["error"] = err
			switch err.(type) {
			case NestableError:
				ne := err.(NestableError)
				if ne.CausedBy((*bvariable.InvalidExpression)(nil)) {
					log.WithFields(logAttrs).Warnln("Invalid Expression to extract")
					values = []string{}
				} else {
					log.WithFields(logAttrs).Errorln("extract error")
					return err
				}
			default:
				log.WithFields(logAttrs).Errorln("extract error")
				return err
			}
		}
		key := strings.Join(values, " ")
		if key == "" {
			key = "default"
		}
		t := job.config.Options[key]
		logAttrs["options_key"] = key
		logAttrs["command_template"] = t
		log.WithFields(logAttrs).Debugln("extracting command of options")
		if t == nil {
			msg := fmt.Sprintf("Invalid command options key: %q", key)
			log.WithFields(logAttrs).Errorln(msg)
			return &InvalidJobError{msg: msg}
		}
		values, err = job.extract(v, t)
		if err != nil {
			logAttrs["error"] = err
			log.WithFields(logAttrs).Errorln("extract error")
			return err
		}
	} else {
		logAttrs["command_template"] = job.config.Template
		if err != nil {
			logAttrs["error"] = err
			log.WithFields(logAttrs).Errorln("extract error")
			return err
		}
	}
	logAttrs["command"] = values
	log.WithFields(logAttrs).Debugln("")
	job.cmd = exec.Command(values[0], values[1:]...)
	job.cmd.Stdout = os.Stdout
	job.cmd.Stderr = os.Stderr
	return nil
}

func (job *Job) extract(v *bvariable.Variable, values []string) ([]string, error) {
	result := []string{}
	errors := []error{}
	for _, src := range values {
		extracted, err := v.Expand(src)
		err = job.convertError(err)
		if err != nil {
			errors = append(errors, &InvalidJobError{cause: err})
			continue
		}
		vals := strings.Split(extracted, v.Separator)
		for _, val := range vals {
			result = append(result, val)
		}
	}
	if len(errors) > 0 {
		return nil, &CompositeError{errors}
	}
	return result, nil
}

func (job *Job) convertError(src error) error {
	switch src.(type) {
	case *bvariable.Errors:
		err := src.(*bvariable.Errors)
		return &CompositeError{[]error(*err)}
	default:
		return src
	}
}

func (job *Job) downloadFiles() error {
	targets := []*Target{}
	for remoteURL, destPath := range job.downloadFileMap {
		url, err := url.Parse(remoteURL)
		if err != nil {
			log.WithFields(log.Fields{"url": remoteURL, "error": err}).Errorln("Invalid URL")
			return err
		}

		dir := path.Dir(destPath)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}

		t := Target{
			Bucket:    url.Host,
			Object:    url.Path[1:],
			LocalPath: destPath,
		}
		targets = append(targets, &t)
		log.WithFields(log.Fields{"target": t}).Debugln("Preparing targets")
	}

	downloaders := TargetWorkers{}
	for i := 0; i < job.config.Downloaders; i++ {
		downloader := &TargetWorker{
			name: "downoad",
			impl: job.storage.Download,
		}
		downloaders = append(downloaders, downloader)
	}
	log.WithFields(log.Fields{"downloaders": len(downloaders)}).Debugln("Downloaders are running")

	log.WithFields(log.Fields{"targets": targets}).Debugln("downloaders processing")
	err := downloaders.process(targets)
	return err
}

func (job *Job) execute() error {
	if job.config.Dryrun {
		return nil
	}
	log.WithFields(log.Fields{"cmd": job.cmd}).Debugln("EXECUTING")
	err := job.cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{"cmd": job.cmd, "error": err}).Errorln("Command returned error")
		return err
	}
	return nil
}

func (job *Job) uploadFiles() error {
	localPaths, err := job.listFiles(job.uploads_dir)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"files": localPaths}).Debugln("Uploading files found")
	targets := []*Target{}
	for _, localPath := range localPaths {
		relPath, err := filepath.Rel(job.uploads_dir, localPath)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "localPath": localPath, "uploads_dir": job.uploads_dir}).Errorln("Failed to get relative path")
			return err
		}
		sep := string([]rune{os.PathSeparator})
		parts := strings.Split(relPath, sep)
		t := Target{
			Bucket:    parts[0],
			Object:    strings.Join(parts[1:], sep),
			LocalPath: localPath,
		}
		targets = append(targets, &t)
		log.WithFields(log.Fields{"target": t}).Debugln("Preparing targets")
	}

	uploaders := TargetWorkers{}
	for i := 0; i < job.config.Uploaders; i++ {
		uploader := &TargetWorker{
			name: "upload",
			impl: job.storage.Upload,
		}
		uploaders = append(uploaders, uploader)
	}
	log.WithFields(log.Fields{"uploaders": len(uploaders)}).Debugln("Uploaders are running")

	err = uploaders.process(targets)
	return err
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
		log.WithFields(log.Fields{"error": err}).Errorln("Error to list upload files")
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
