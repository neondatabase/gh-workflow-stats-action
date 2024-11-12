package db

import (
	"context"

	"github.com/google/go-github/v65/github"
	"github.com/jmoiron/sqlx"

	"github.com/neondatabase/gh-workflow-stats-action/internal/githubclient"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
)

type Repository interface {
	SaveWorkflowRun(record *data.WorkflowRunRec) error
	SaveWorkflowRunAttempt(workflowRun *github.WorkflowRun) error
	InsertJob(tx *sqlx.Tx, workflowJob *github.WorkflowJob) error
	InsertSteps(tx *sqlx.Tx, workflowJob *github.WorkflowJob) error
	WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) error
	QueryWorkflowRunAttempts(runId int64) map[int64]struct{}
	QueryWorkflowRunsNotInDb(workflowRuns []githubclient.WorkflowRunAttemptKey) []githubclient.WorkflowRunAttemptKey
}
