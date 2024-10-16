package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v65/github"
)

func main() {
	ctx := context.Background()

	// Get env vars
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = connectDB(&conf)
	if err != nil {
		log.Fatal(err)
	}
	err = initDatabase(conf)
	if err != nil {
		log.Fatal(err)
	}

	initGhClient(&conf)

	var workflowStat *WorkflowRunRec
	workflowStat, err = getWorkflowStat(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	err = saveWorkflowRun(conf, workflowStat)
	if err != nil {
		log.Fatal(err)
	}

	var lastAttemptRun *github.WorkflowRun
	lastAttemptN := workflowStat.RunAttempt
	lastAttemptRun, err = getWorkflowAttempt(ctx, conf, lastAttemptN)
	if err != nil {
		log.Fatal(err)
	}
	err = saveWorkflowRunAttempt(conf, lastAttemptRun)
	if err != nil {
		log.Fatal(err)
	}

	jobsInfo, err := getWorkflowAttemptJobs(ctx, conf, lastAttemptN)
	if err != nil {
		log.Fatal(err)
	}
	var dbContext dbContextType
	prepareJobTransaction(ctx, conf, &dbContext)
	for _, jobInfo := range jobsInfo {
		err = saveJobInfo(&dbContext, jobInfo)
		if err != nil {
			fmt.Println(err)
		}
	}
	err = commitJobTransaction(&dbContext)
	if err != nil {
		fmt.Println(err)
	}
}
