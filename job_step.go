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
	PREPARING   JobStep = 1 + iota
	DOWNLOADING
	EXECUTING
	UPLOADING
	CANCELLING
	ACKSENDING
	NACKSENDING
	CLEANUP
)

var (
	JOB_STEP_DEFS = map[JobStep][]string {
		PREPARING:    []string{"PREPARING"	, "info" , "error"},
		DOWNLOADING:	[]string{"DOWNLOADING", "info" , "error"},
		EXECUTING:		[]string{"EXECUTING"	,	"info" , "error"},
		UPLOADING:		[]string{"UPLOADING"	,	"info" , "error"},
		CANCELLING:		[]string{"CANCELLING" , "info" , "fatal"},
		ACKSENDING:		[]string{"ACKSENDING" , "info" , "error"},
		NACKSENDING:	[]string{"NACKSENDING", "info" , "warn" },
		CLEANUP:			[]string{"CLEANUP"		, "debug", "warn" },
	}
)

func (js JobStep) String() string {
	return JOB_STEP_DEFS[js][0]
}
func (js JobStep) successLogLevel() string {
	return JOB_STEP_DEFS[js][1]
}
func (js JobStep) failureLogLevel() string {
	return JOB_STEP_DEFS[js][2]
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
