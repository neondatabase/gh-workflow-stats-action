package main

import (
	"context"
	"log"
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

	// lastAttemptRun, err = getWorkflowAttempt(ctx, conf, workflowStat.RunAttempt)
}
