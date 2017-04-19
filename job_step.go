package main

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
		successLogLevel string
		failureLogLevel string
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
		INITIALIZING: JobStepDef{"INITIALIZING", "info", "error", PREPARING},
		DOWNLOADING:  JobStepDef{"DOWNLOADING", "debug", "error", WORKING},
		EXECUTING:    JobStepDef{"EXECUTING", "debug", "error", WORKING},
		UPLOADING:    JobStepDef{"UPLOADING", "debug", "error", WORKING},
		CLEANUP:      JobStepDef{"CLEANUP", "debug", "warn", WORKING},
		NACKSENDING:  JobStepDef{"NACKSENDING", "warn", "error", RETRYING},
		CANCELLING:   JobStepDef{"CANCELLING", "error", "fatal", INVALID_JOB},
		ACKSENDING:   JobStepDef{"ACKSENDING", "info", "fatal", COMPLETED},
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js].name
}
func (js JobStep) successLogLevel() string {
	return JOB_STEP_DEFS[js].successLogLevel
}
func (js JobStep) failureLogLevel() string {
	return JOB_STEP_DEFS[js].failureLogLevel
}
func (js JobStep) baseProgress() Progress {
	return JOB_STEP_DEFS[js].baseProgress
}

func (js JobStep) completed(st JobStepStatus) bool {
	return (js == ACKSENDING) && (st == SUCCESS)
}
func (js JobStep) logLevelFor(st JobStepStatus) string {
	switch st {
	case STARTING:
		return "debug"
	case SUCCESS:
		return js.successLogLevel()
	case FAILURE:
		return js.failureLogLevel()
	default:
		return "info"
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
