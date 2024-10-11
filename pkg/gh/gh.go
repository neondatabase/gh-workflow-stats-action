package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"

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
	var token *http.Client
	if len(conf.GithubToken) != 0 {
		token = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: conf.GithubToken},
		))
	}

	conf.GhClient = github.NewClient(token)
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

func GetWorkflowAttemptJobs(ctx context.Context, conf config.ConfigType, attempt int64) ([]*github.WorkflowJob, error) {
	var result []*github.WorkflowJob

	opts := &github.ListOptions{PerPage: 100}
	for {
		jobsData, resp, err := conf.GhClient.Actions.ListWorkflowJobsAttempt(
			ctx,
			conf.Owner, conf.Repo,
			conf.RunID,
			attempt,
			opts,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, jobsData.Jobs...)
		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}
	return result, nil
}
