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

type WorkflowJobRec struct {
	JobId           int64
	RunId           int64
	NodeID          string
	HeadBranch      string
	HeadSHA         string
	Status          string
	Conclusion      string
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time
	Name            string
	RunnerName      string
	RunnerGroupName string
	RunAttempt      int64
	WorkflowName    string
}

func ghWorkflowJobRec(j *github.WorkflowJob) *WorkflowJobRec {
	return &WorkflowJobRec{
		JobId:           j.GetID(),
		RunId:           j.GetRunID(),
		NodeID:          j.GetNodeID(),
		HeadBranch:      j.GetHeadBranch(),
		HeadSHA:         j.GetHeadSHA(),
		Status:          j.GetStatus(),
		Conclusion:      j.GetConclusion(),
		CreatedAt:       j.GetCreatedAt().Time,
		StartedAt:       j.GetStartedAt().Time,
		CompletedAt:     j.GetCompletedAt().Time,
		Name:            j.GetName(),
		RunnerName:      j.GetRunnerName(),
		RunnerGroupName: j.GetRunnerGroupName(),
		RunAttempt:      j.GetRunAttempt(),
		WorkflowName:    j.GetWorkflowName(),
	}
}

type WorkflowJobStepRec struct {
	JobId       int64
	RunId       int64
	RunAttempt  int64
	Name        string
	Status      string
	Conclusion  string
	Number      int64
	StartedAt   time.Time
	CompletedAt time.Time
}

func ghWorkflowJobStepRec(j *github.WorkflowJob, s *github.TaskStep) *WorkflowJobStepRec {
	return &WorkflowJobStepRec{
		JobId:       j.GetID(),
		RunId:       j.GetRunID(),
		RunAttempt:  j.GetRunAttempt(),
		Name:        s.GetName(),
		Status:      s.GetStatus(),
		Conclusion:  s.GetConclusion(),
		Number:      s.GetNumber(),
		StartedAt:   s.GetStartedAt().Time,
		CompletedAt: s.GetCompletedAt().Time,
	}
}
