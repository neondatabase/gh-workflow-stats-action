package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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

func createRecords(ctx context.Context, conf configType) (*WorkflowStat, error) {
	var token *http.Client
	if len(conf.githubToken) != 0 {
		token = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: conf.githubToken},
		))
	}

	client := github.NewClient(token)

	runID, err := strconv.ParseInt(conf.runID, 10, 64)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Getting data for %s/%s, runId %d\n", conf.owner, conf.repo, runID)
	workflowRunData, _, err := client.Actions.GetWorkflowRunByID(ctx, conf.owner, conf.repo, runID)
	if err != nil {
		return nil, err
	}

	if workflowRunData == nil {
		fmt.Printf("Got nil\n")
		return &WorkflowStat{RepoName: conf.repository}, nil
	}

	attemptData, _, err := client.Actions.GetWorkflowRunAttempt(
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

	jobsData, _, err := client.Actions.ListWorkflowJobsAttempt(
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

	return &WorkflowStat{
		WorkflowId: workflowRunData.GetWorkflowID(),
		Name:       workflowRunData.GetName(),
		Status:     workflowRunData.GetStatus(),
		Conclusion: workflowRunData.GetConclusion(),
		RunId:      workflowRunData.GetID(),
		RunAttempt: int64(workflowRunData.GetRunAttempt()),
		StartedAt:  workflowRunData.GetCreatedAt().Time,
		UpdatedAt:  workflowRunData.GetUpdatedAt().Time,
		RepoName:   workflowRunData.GetRepository().GetFullName(),
		Event:      workflowRunData.GetEvent(),
	}, nil
}
