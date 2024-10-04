package main

import (
	"time"

	"github.com/google/go-github/v65/github"
)

type WorkflowRunRec struct {
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

func ghWorkflowRunRec(w *github.WorkflowRun) *WorkflowRunRec {
	return &WorkflowRunRec{
		WorkflowId: w.GetWorkflowID(),
		Name:       w.GetName(),
		Status:     w.GetStatus(),
		Conclusion: w.GetConclusion(),
		RunId:      w.GetID(),
		RunAttempt: int64(w.GetRunAttempt()),
		StartedAt:  w.GetCreatedAt().Time,
		UpdatedAt:  w.GetUpdatedAt().Time,
		RepoName:   w.GetRepository().GetFullName(),
		Event:      w.GetEvent(),
	}
}
