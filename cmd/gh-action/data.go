package main

import (
	"time"
)

type WorkflowStat struct {
	WorkflowId int64
	Name       string
	Status     string
	Conclusion string
	RunId      int64
	RunAttempt int64
	StartedAt  time.Time
	UpdatedAt  time.Time
	RepoName   string
	Event      string
}
