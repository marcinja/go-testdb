package main

import "time"

// TODO: Think about which level to include logic about version numbers/RCs.

type Status int

const (
	PASSED = iota
	SKIPPED
	FAILED
	UNDETERMINED
)

var StatusStrings = [...]string{"PASSED", "SKIPPED", "FAILED", "UNDETERMINED"}

type Result struct {
	commitHash    string
	dateTime      time.Time
	testResults   []*TestResult
	moduleResults []*ModuleResult
}

// ModuleResult stores information about the tests of a single module.
type ModuleResult struct {
	name     string
	result   Status
	duration time.Duration
}

// TestResult represents the information given from a single test completing.
type TestResult struct {
	name     string
	result   Status
	output   string
	duration time.Duration
}
