package main

import (
	"time"
)

type WorkflowStat struct {
	WorkflowId int64
	Name       string
	Status     string
	Conclusion string
	RunId      int
	RunAttempt int
	StartedAt  time.Time
	UpdatedAt  time.Time
	RepoName   string
	Event      string
}
