package main

type JobStepStatus int

const (
	STARTING JobStepStatus = 1 + iota
	SUCCESS
	FAILURE
)

func (jss JobStepStatus) String() string {
	switch jss {
	case STARTING: return "STARTING"
	case SUCCESS: return "SUCCESS"
	case FAILURE: return "FAILURE"
	default: return "Unknown"
	}
}

type JobStep int

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
	JOB_STEP_DEFS = map[JobStep][]string {
		INITIALIZING: []string{"INITIALIZING"	, "error"},
		DOWNLOADING:	[]string{"DOWNLOADING", "error"},
		EXECUTING:		[]string{"EXECUTING"	,	"error"},
		UPLOADING:		[]string{"UPLOADING"	,	"error"},
		CLEANUP:			[]string{"CLEANUP"		, "warn" },
		NACKSENDING:	[]string{"NACKSENDING", "warn" },
		CANCELLING:		[]string{"CANCELLING" , "fatal"},
		ACKSENDING:		[]string{"ACKSENDING" , "fatal"},
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js][0]
}
func (js JobStep) failureLogLevel() string {
	return JOB_STEP_DEFS[js][1]
}

func (js JobStep) completed(st JobStepStatus) bool {
	return (js == ACKSENDING) && (st == SUCCESS)
}
func (js JobStep) logLevelFor(st JobStepStatus) string {
	switch st {
	// case STARTING: return "info"
	case SUCCESS: return js.successLogLevel()
	case FAILURE: return js.failureLogLevel()
	default: return "info"
	}
}
