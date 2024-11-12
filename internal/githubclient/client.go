package githubclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
)

type ClientImpl struct {
	cfg   *Config
	owner string
	repo  string

	client *github.Client
}

func NewClient(cfg *Config) (*ClientImpl, error) {
	repoDetails := strings.Split(cfg.Repository, "/")
	if len(repoDetails) != 2 {
		return nil, fmt.Errorf("invalid config: GITHUB_REPOSITORY")
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	if cfg.MaxRetry != 0 {
		retryClient.RetryMax = cfg.MaxRetry
	}

	rl, err := github_ratelimit.NewRateLimitWaiterClient(retryClient.StandardClient().Transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limit waiter: %v", err)
	}

	return &ClientImpl{
		cfg:    cfg,
		owner:  repoDetails[0],
		repo:   repoDetails[1],
		client: github.NewClient(rl).WithAuthToken(cfg.Token),
	}, nil
}

func (c *ClientImpl) GetWorkflowStat(ctx context.Context, runID int64) (*data.WorkflowRunRec, error) {
	fmt.Printf("Getting data for %s/%s, runID %d\n", c.owner, c.repo, runID)
	workflowRunData, _, err := c.client.Actions.GetWorkflowRunByID(ctx, c.owner, c.repo, runID)
	if err != nil {
		return nil, err
	}

	if workflowRunData == nil {
		fmt.Printf("Got nil\n")
		return &data.WorkflowRunRec{RepoName: c.repo}, nil
	}

	return data.GhWorkflowRunRec(workflowRunData), nil
}

func (c *ClientImpl) GetWorkflowAttempt(ctx context.Context, runID int64, attempt int64) (*github.WorkflowRun, error) {
	workflowRunData, _, err := c.client.Actions.GetWorkflowRunAttempt(
		ctx,
		c.owner, c.repo,
		runID,
		int(attempt),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return workflowRunData, nil
}

func (c *ClientImpl) GetWorkflowAttemptJobs(ctx context.Context, runID int64, attempt int64) ([]*github.WorkflowJob, github.Rate, error) {
	var result []*github.WorkflowJob
	finalRate := github.Rate{}

	opts := &github.ListOptions{PerPage: 100}
	for {
		jobsData, resp, err := c.client.Actions.ListWorkflowJobsAttempt(
			ctx,
			c.owner, c.repo,
			runID,
			attempt,
			opts,
		)
		if resp != nil {
			finalRate = resp.Rate
		}
		if err != nil {
			return nil, finalRate, err
		}
		result = append(result, jobsData.Jobs...)
		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}
	return result, finalRate, nil
}

type WorkflowRunAttemptKey struct {
	RunId      int64
	RunAttempt int64
}

func (c *ClientImpl) ListWorkflowRuns(
	ctx context.Context,
	start time.Time, end time.Time,
) (map[WorkflowRunAttemptKey]*github.WorkflowRun, github.Rate, error) {
	result := make(map[WorkflowRunAttemptKey]*github.WorkflowRun)
	finalRate := github.Rate{}

	opts := &github.ListOptions{PerPage: 100}
	for {
		workflowRuns, resp, err := c.client.Actions.ListRepositoryWorkflowRuns(
			ctx,
			c.owner, c.repo,
			&github.ListWorkflowRunsOptions{
				Created:     fmt.Sprintf("%s..%s", start.Format(time.RFC3339), end.Format(time.RFC3339)),
				Status:      "completed",
				ListOptions: *opts,
			},
		)
		if resp != nil {
			finalRate = resp.Rate
		}
		if err != nil {
			return nil, finalRate, err
		}
		for _, rec := range workflowRuns.WorkflowRuns {
			key := WorkflowRunAttemptKey{RunId: rec.GetID(), RunAttempt: int64(rec.GetRunAttempt())}
			if v, ok := result[key]; ok {
				fmt.Printf("Strange, record is already stored for %v (%+v), updating with %+v\n", key, v, rec)
			}
			result[key] = rec
		}
		if resp.NextPage == 0 {
			finalRate = resp.Rate
			break
		}
		opts.Page = resp.NextPage
	}
	return result, finalRate, nil
}
