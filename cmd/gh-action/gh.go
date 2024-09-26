package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
)

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
	workflowData, _, err := client.Actions.GetWorkflowRunByID(ctx, conf.owner, conf.repo, runID)
	if err != nil {
		return nil, err
	}

	if workflowData == nil {
		fmt.Printf("Got nil\n")
		return &WorkflowStat{RepoName: conf.repository}, nil
	}

	return &WorkflowStat{
		WorkflowId: *workflowData.ID,
		Name:       *workflowData.Name,
		Status:     *workflowData.Status,
		Conclusion: *workflowData.Conclusion,
		RunId:      *workflowData.RunNumber,
		RunAttempt: *workflowData.RunAttempt,
		StartedAt:  workflowData.CreatedAt.Time,
		UpdatedAt:  workflowData.UpdatedAt.Time,
		RepoName:   *workflowData.Repository.FullName,
		Event:      *workflowData.Event,
	}, nil
}
