package main

import (
	"bytes"
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
	"time"

	"github.com/groovenauts/blocks-variable"
	"github.com/groovenauts/concurrent-go"
	"github.com/satori/go.uuid"

	logrus "github.com/sirupsen/logrus"
)

type Job struct {
	config *CommandConfig

	commandSeverityLevel logrus.Level // From LogConfig

	downloadConfig *DownloadConfig
	uploadConfig   *UploadConfig

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

	// This is set at setupExecUUID
	execUUID string

	outputBuffer *bytes.Buffer

	cmd *exec.Cmd
}

const (
	StartTimeKey  = "job.start-time"
	FinishTimeKey = "job.finish-time"
)

func (job *Job) run() error {
	job.message.raw.Message.Attributes[StartTimeKey] = time.Now().Format(time.RFC3339)
	err := job.runWithoutErrorHandling()
	job.message.raw.Message.Attributes[FinishTimeKey] = time.Now().Format(time.RFC3339)
	step := map[bool]JobStep{false: ACKSENDING, true: CANCELLING}[err != nil]
	e := job.withNotify(step, job.message.Ack)()
	if e != nil {
		return e
	}
	return nil
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

	return nil
}

func (job *Job) withNotify(step JobStep, f func() error) func() error {
	return job.notification.wrap(job.message.MessageId(), step, job.message.raw.Message.Attributes, f)
}

func (job *Job) prepare() error {
	log := log.WithFields(logrus.Fields{"job_message_id": job.message.MessageId()})
	err := job.message.Validate()
	if err != nil {
		logAttrs := logrus.Fields{
			"message": job.message.raw.Message,
			"error":   err,
		}
		log.WithFields(logAttrs).Errorf("Invalid Message")
		return err
	}

	job.message.InsertExecUUID()

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
		logAttrs := logrus.Fields{
			"template": job.config.Template,
			"message":  job.message,
			"error":    err,
		}
		log.WithFields(logAttrs).Errorf("Failed to build command")
		return err
	}
	return nil
}

func (job *Job) setupExecUUID() {
	job.execUUID = uuid.NewV4().String()
	job.message.raw.Message.Attributes[ExecUUIDKey] = job.execUUID
}

func (job *Job) setupWorkspace() error {
	dir := job.workspace
	if dir == "" {
		var err error
		dir, err = ioutil.TempDir("", "workspace")
		if err != nil {
			log.Fatal(err)
			return err
		}
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
		logAttrs := logrus.Fields{"error": err, "data": job.message.raw.Message.Data}
		log.WithFields(logAttrs).Errorf("Failed to decode by base64")
		return err
	}
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(decoded), &parsed)
	if err != nil {
		logAttrs := logrus.Fields{"error": err, "data": job.message.raw.Message.Data}
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
				logAttrs := logrus.Fields{"error": err, "obj": obj}
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
			log.WithFields(logrus.Fields{"url": obj}).Errorf("Invalid download file URL: %T\n", obj)
		}
	}
	for _, remote_url := range remoteUrls {
		url, err := url.Parse(remote_url)
		if err != nil {
			log.WithFields(logrus.Fields{"url": remote_url}).Errorln("Invalid download file URL")
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
	v := job.buildVariable()
	values, err := job.extract(v, job.config.Template)
	if len(job.config.Options) > 0 {
		log := log.WithFields(logrus.Fields{
			"options_key_template": job.config.Template,
			"options_key_base":     values,
		})
		log.Debugln("extracting key of options")
		if err != nil {
			log = log.WithFields(logrus.Fields{"error": err})
			switch err.(type) {
			case NestableError:
				ne := err.(NestableError)
				if ne.CausedBy((*bvariable.InvalidExpression)(nil)) {
					log.Warnln("Invalid Expression to extract")
					values = []string{}
				} else {
					log.Errorln("extract error")
					return err
				}
			default:
				log.Errorln("extract error")
				return err
			}
		}
		key := strings.Join(values, " ")
		if key == "" {
			key = "default"
		}
		t := job.config.Options[key]
		log = log.WithFields(logrus.Fields{
			"options_key":      key,
			"command_template": t,
		})
		log.Debugln("extracting command of options")
		if t == nil {
			msg := fmt.Sprintf("Invalid command options key: %q", key)
			log.Errorln(msg)
			return &InvalidJobError{msg: msg}
		}
		values, err = job.extract(v, t)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err}).Errorln("extract error")
			return err
		}
	} else {
		log := log.WithFields(logrus.Fields{"command_template": job.config.Template})
		if err != nil {
			log.WithFields(logrus.Fields{"error": err}).Errorln("extract error")
			return err
		}
	}
	job.outputBuffer = &bytes.Buffer{}
	w := &LogrusWriter{Dest: log, Severity: job.commandSeverityLevel}
	w.Setup()
	out := &CompositeWriter{Main: job.outputBuffer, Sub: w}
	cmd := exec.Command(values[0], values[1:]...)
	cmd.Stdout = out
	cmd.Stderr = out
	job.cmd = cmd
	log.WithFields(logrus.Fields{"job.cmd": job.cmd}).Debugln("Job#build has done")
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
			log.WithFields(logrus.Fields{"url": remoteURL, "error": err}).Errorln("Invalid URL")
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
	}
	log.WithFields(logrus.Fields{"targets": targets}).Debugln("Download Prepared")

	jobs := concurrent.Jobs{}
	for _, target := range targets {
		jobs = append(jobs, &concurrent.Job{Payload: target})
	}

	rf := &RetryableFunc{
		name:     "downoad",
		maxTries: job.downloadConfig.Worker.MaxTries,
		interval: 30 * time.Second,
	}
	f := rf.WithLog(rf.Wrap(func(j *concurrent.Job) error {
		t, ok := j.Payload.(*Target)
		if !ok {
			return fmt.Errorf("Unknown Payload: %v\n", j.Payload)
		}
		return job.storage.Download(t.Bucket, t.Object, t.LocalPath)
	}))

	downloaders := concurrent.NewWorkers(f, job.downloadConfig.Worker.Workers)
	log.WithFields(logrus.Fields{"downloaders": len(downloaders)}).Debugln("Downloaders are running")

	log.WithFields(logrus.Fields{"targets": targets}).Debugln("downloaders processing")
	downloaders.Process(jobs)
	return jobs.Error()
}

func (job *Job) execute() error {
	if job.config.Dryrun {
		return nil
	}
	log := log.WithFields(logrus.Fields{"cmd": job.cmd})
	log.Debugln("EXECUTING")
	err := job.cmd.Run()
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Errorln("Command returned error")
		return fmt.Errorf("[%T] %v\noutput:\n%s", err, err.Error(), job.outputBuffer.String())
	}
	return nil
}

func (job *Job) uploadFiles() error {
	localPaths, err := job.listFiles(job.uploads_dir)
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{"files": localPaths}).Debugln("Uploading files found")
	targets := []*Target{}
	for _, localPath := range localPaths {
		relPath, err := filepath.Rel(job.uploads_dir, localPath)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err, "localPath": localPath, "uploads_dir": job.uploads_dir}).Errorln("Failed to get relative path")
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
	}
	log.WithFields(logrus.Fields{"targets": targets}).Debugln("Upload Prepared")

	jobs := concurrent.Jobs{}
	for _, target := range targets {
		jobs = append(jobs, &concurrent.Job{Payload: target})
	}

	rf := &RetryableFunc{
		name:     "upload",
		maxTries: job.uploadConfig.Worker.MaxTries,
		interval: 30 * time.Second,
	}
	f := rf.WithLog(rf.Wrap(func(j *concurrent.Job) error {
		t, ok := j.Payload.(*Target)
		if !ok {
			return fmt.Errorf("Unknown Payload: %v\n", j.Payload)
		}
		return job.storage.Upload(t.Bucket, t.Object, t.LocalPath)
	}))

	uploaders := concurrent.NewWorkers(f, job.uploadConfig.Worker.Workers)
	log.WithFields(logrus.Fields{"uploaders": len(uploaders)}).Debugln("Uploaders are running")

	uploaders.Process(jobs)
	return jobs.Error()
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
		log.WithFields(logrus.Fields{"error": err}).Errorln("Error to list upload files")
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
