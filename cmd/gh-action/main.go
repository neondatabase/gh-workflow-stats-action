package main

// Trigger action

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v65/github"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/db"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/gh"
)

func main() {
	ctx := context.Background()

	// Get env vars
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = db.ConnectDB(&conf)
	if err != nil {
		log.Fatal(err)
	}
	err = db.InitDatabase(conf)
	if err != nil {
		log.Fatal(err)
	}

	gh.InitGhClient(&conf)

	var workflowStat *data.WorkflowRunRec
	workflowStat, err = gh.GetWorkflowStat(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	err = db.SaveWorkflowRun(conf, workflowStat)
	if err != nil {
		log.Fatal(err)
	}

	var lastAttemptRun *github.WorkflowRun
	lastAttemptN := workflowStat.RunAttempt
	lastAttemptRun, err = gh.GetWorkflowAttempt(ctx, conf, lastAttemptN)
	if err != nil {
		log.Fatal(err)
	}
	err = db.SaveWorkflowRunAttempt(conf, lastAttemptRun)
	if err != nil {
		log.Fatal(err)
	}

	jobsInfo, _, err := gh.GetWorkflowAttemptJobs(ctx, conf, lastAttemptN)
	if err != nil {
		log.Fatal(err)
	}
	var dbContext config.DbContextType
	db.PrepareJobTransaction(ctx, conf, &dbContext)
	for _, jobInfo := range jobsInfo {
		err = db.SaveJobInfo(&dbContext, jobInfo)
		if err != nil {
			fmt.Println(err)
		}
	}
	err = db.CommitJobTransaction(&dbContext)
	if err != nil {
		fmt.Println(err)
	}
}
