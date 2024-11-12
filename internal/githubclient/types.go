package githubclient

import (
	"context"
	"time"

	"github.com/google/go-github/v65/github"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
)

type Client interface {
	GetWorkflowStat(ctx context.Context, runID int64) (*data.WorkflowRunRec, error)
	GetWorkflowAttempt(ctx context.Context, runID int64, attempt int64) (*github.WorkflowRun, error)
	GetWorkflowAttemptJobs(ctx context.Context, runID int64, attempt int64) ([]*github.WorkflowJob, github.Rate, error)
	ListWorkflowRuns(
		ctx context.Context,
		start time.Time, end time.Time,
	) (map[WorkflowRunAttemptKey]*github.WorkflowRun, github.Rate, error)
}
