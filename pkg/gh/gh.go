package gh

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v65/github"
	"github.com/hashicorp/go-retryablehttp"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
)

func printJobInfo(job *github.WorkflowJob) {
	fmt.Printf("== Job %s %s, (created: %v, started: %v, completed: %v)\n",
		*job.Name,
		*job.Status,
		*job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	)
	for _, step := range job.Steps {
		fmt.Printf("Step %s, started %v, completed %v\n", *step.Name, step.StartedAt, step.CompletedAt)
	}
}

func InitGhClient(conf *config.ConfigType) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	rl, err := github_ratelimit.NewRateLimitWaiterClient(retryClient.StandardClient().Transport)
	if err != nil {
		log.Fatal(err)
	}
	conf.GhClient = github.NewClient(rl).WithAuthToken(conf.GithubToken)
}

func GetWorkflowStat(ctx context.Context, conf config.ConfigType) (*data.WorkflowRunRec, error) {
	fmt.Printf("Getting data for %s/%s, runId %d\n", conf.Owner, conf.Repo, conf.RunID)
	workflowRunData, _, err := conf.GhClient.Actions.GetWorkflowRunByID(ctx, conf.Owner, conf.Repo, conf.RunID)
	if err != nil {
		return nil, err
	}

	if workflowRunData == nil {
		fmt.Printf("Got nil\n")
		return &data.WorkflowRunRec{RepoName: conf.Repository}, nil
	}

	return data.GhWorkflowRunRec(workflowRunData), nil
}

func GetWorkflowAttempt(ctx context.Context, conf config.ConfigType, attempt int64) (*github.WorkflowRun, error) {
	workflowRunData, _, err := conf.GhClient.Actions.GetWorkflowRunAttempt(
		ctx,
		conf.Owner, conf.Repo,
		conf.RunID,
		int(attempt),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return workflowRunData, nil
}

func GetWorkflowAttemptJobs(ctx context.Context, conf config.ConfigType, attempt int64) ([]*github.WorkflowJob, github.Rate, error) {
	var result []*github.WorkflowJob
	finalRate := github.Rate{}

	opts := &github.ListOptions{PerPage: 100}
	for {
		jobsData, resp, err := conf.GhClient.Actions.ListWorkflowJobsAttempt(
			ctx,
			conf.Owner, conf.Repo,
			conf.RunID,
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

func ListWorkflowRuns(ctx context.Context,
	conf config.ConfigType,
	start time.Time, end time.Time) (map[WorkflowRunAttemptKey]*github.WorkflowRun, github.Rate, error) {
	result := make(map[WorkflowRunAttemptKey]*github.WorkflowRun)
	finalRate := github.Rate{}

	opts := &github.ListOptions{PerPage: 100}
	for {
		workflowRuns, resp, err := conf.GhClient.Actions.ListRepositoryWorkflowRuns(
			ctx,
			conf.Owner, conf.Repo,
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
