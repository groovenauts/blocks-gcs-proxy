package main

import (
	logrus "github.com/sirupsen/logrus"
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
		successLogLevel logrus.Level
		failureLogLevel logrus.Level
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
		INITIALIZING: JobStepDef{"INITIALIZING", logrus.InfoLevel, logrus.ErrorLevel, PREPARING},
		DOWNLOADING:  JobStepDef{"DOWNLOADING", logrus.DebugLevel, logrus.ErrorLevel, WORKING},
		EXECUTING:    JobStepDef{"EXECUTING", logrus.DebugLevel, logrus.ErrorLevel, WORKING},
		UPLOADING:    JobStepDef{"UPLOADING", logrus.DebugLevel, logrus.ErrorLevel, WORKING},
		CLEANUP:      JobStepDef{"CLEANUP", logrus.DebugLevel, logrus.WarnLevel, WORKING},
		NACKSENDING:  JobStepDef{"NACKSENDING", logrus.WarnLevel, logrus.ErrorLevel, RETRYING},
		CANCELLING:   JobStepDef{"CANCELLING", logrus.ErrorLevel, logrus.FatalLevel, INVALID_JOB},
		ACKSENDING:   JobStepDef{"ACKSENDING", logrus.InfoLevel, logrus.FatalLevel, COMPLETED},
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js].name
}
func (js JobStep) successLogLevel() logrus.Level {
	return JOB_STEP_DEFS[js].successLogLevel
}
func (js JobStep) failureLogLevel() logrus.Level {
	return JOB_STEP_DEFS[js].failureLogLevel
}
func (js JobStep) baseProgress() Progress {
	return JOB_STEP_DEFS[js].baseProgress
}

func (js JobStep) completed(st JobStepStatus) bool {
	return (js == ACKSENDING) && (st == SUCCESS)
}
func (js JobStep) logLevelFor(st JobStepStatus) logrus.Level {
	switch st {
	case STARTING:
		return logrus.DebugLevel
	case SUCCESS:
		return js.successLogLevel()
	case FAILURE:
		return js.failureLogLevel()
	default:
		return logrus.InfoLevel
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
