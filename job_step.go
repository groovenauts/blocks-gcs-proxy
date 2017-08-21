package main

import (
	log "github.com/sirupsen/logrus"
)

type JobStepStatus int

const (
	STARTING JobStepStatus = 1 + iota
	SUCCESS
	FAILURE
)

func (jss JobStepStatus) String() string {
	switch jss {
	case STARTING:
		return "STARTING"
	case SUCCESS:
		return "SUCCESS"
	case FAILURE:
		return "FAILURE"
	default:
		return "Unknown"
	}
}

type (
	JobStep    int
	JobStepDef struct {
		name            string
		successLogLevel log.Level
		failureLogLevel log.Level
		baseProgress    Progress
	}
)

const (
	INITIALIZING JobStep = 1 + iota
	DOWNLOADING
	EXECUTING
	UPLOADING
	CLEANUP
	NACKSENDING
	CANCELLING
	ACKSENDING
)

var (
	JOB_STEP_DEFS = map[JobStep]JobStepDef{
		INITIALIZING: JobStepDef{"INITIALIZING", log.InfoLevel, log.ErrorLevel, PREPARING},
		DOWNLOADING:  JobStepDef{"DOWNLOADING", log.DebugLevel, log.ErrorLevel, WORKING},
		EXECUTING:    JobStepDef{"EXECUTING", log.DebugLevel, log.ErrorLevel, WORKING},
		UPLOADING:    JobStepDef{"UPLOADING", log.DebugLevel, log.ErrorLevel, WORKING},
		CLEANUP:      JobStepDef{"CLEANUP", log.DebugLevel, log.WarnLevel, WORKING},
		NACKSENDING:  JobStepDef{"NACKSENDING", log.WarnLevel, log.ErrorLevel, RETRYING},
		CANCELLING:   JobStepDef{"CANCELLING", log.ErrorLevel, log.FatalLevel, INVALID_JOB},
		ACKSENDING:   JobStepDef{"ACKSENDING", log.InfoLevel, log.FatalLevel, COMPLETED},
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js].name
}
func (js JobStep) successLogLevel() log.Level {
	return JOB_STEP_DEFS[js].successLogLevel
}
func (js JobStep) failureLogLevel() log.Level {
	return JOB_STEP_DEFS[js].failureLogLevel
}
func (js JobStep) baseProgress() Progress {
	return JOB_STEP_DEFS[js].baseProgress
}

func (js JobStep) completed(st JobStepStatus) bool {
	return (js == ACKSENDING) && (st == SUCCESS)
}
func (js JobStep) logLevelFor(st JobStepStatus) log.Level {
	switch st {
	case STARTING:
		return log.DebugLevel
	case SUCCESS:
		return js.successLogLevel()
	case FAILURE:
		return js.failureLogLevel()
	default:
		return log.InfoLevel
	}
}
func (js JobStep) progressFor(st JobStepStatus) Progress {
	switch js {
	case INITIALIZING, DOWNLOADING, EXECUTING, UPLOADING, CLEANUP:
		return js.baseProgress()
	case NACKSENDING, CANCELLING, ACKSENDING:
		switch st {
		case STARTING:
			return WORKING
		case SUCCESS:
			return js.baseProgress()
		case FAILURE:
			return WORKING
		}
	}
	return Progress(0)
}
