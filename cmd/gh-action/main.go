package main

import (
	"context"
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
	lastAttemptRun, err = getWorkflowAttempt(ctx, conf, workflowStat.RunAttempt)
	if err != nil {
		log.Fatal(err)
	}
	saveWorkflowRunAttempt(conf, lastAttemptRun)
}
