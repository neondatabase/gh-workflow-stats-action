package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
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

func initGhClient(conf *configType) {
	var token *http.Client
	if len(conf.githubToken) != 0 {
		token = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: conf.githubToken},
		))
	}

	conf.ghClient = github.NewClient(token)
}

func getWorkflowStat(ctx context.Context, conf configType) (*WorkflowRunRec, error) {
	fmt.Printf("Getting data for %s/%s, runId %d\n", conf.owner, conf.repo, conf.runID)
	workflowRunData, _, err := conf.ghClient.Actions.GetWorkflowRunByID(ctx, conf.owner, conf.repo, conf.runID)
	if err != nil {
		return nil, err
	}

	if workflowRunData == nil {
		fmt.Printf("Got nil\n")
		return &WorkflowRunRec{RepoName: conf.repository}, nil
	}

	attemptData, _, err := conf.ghClient.Actions.GetWorkflowRunAttempt(
		ctx,
		conf.owner, conf.repo,
		*workflowRunData.ID,
		*workflowRunData.RunAttempt,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("AttemptData: %+v\n", attemptData)

	jobsData, _, err := conf.ghClient.Actions.ListWorkflowJobsAttempt(
		ctx,
		conf.owner, conf.repo,
		*attemptData.ID,
		int64(workflowRunData.GetRunAttempt()),
		nil,
	)
	if err != nil {
		return nil, err
	}
	fmt.Printf("JobsData: %+v\n", jobsData)
	for _, job := range jobsData.Jobs {
		printJobInfo(job)
	}

	return ghWorkflowRunRec(workflowRunData), nil
}

func getWorkflowAttempt(ctx context.Context, conf configType, attempt int64) (*github.WorkflowRun, error) {
	workflowRunData, _, err := conf.ghClient.Actions.GetWorkflowRunAttempt(
		ctx,
		conf.owner, conf.repo,
		conf.runID,
		int(attempt),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return workflowRunData, nil
}
