package main

type Progress int

const (
	PREPARING Progress = 1 + iota
	WORKING
	RETRYING
	INVALID_JOB
	COMPLETED
)
