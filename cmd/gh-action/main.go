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

	var workflowStat *WorkflowStat
	workflowStat, err = getWorkflowStat(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	err = saveWorkflowRun(conf, workflowStat)
	if err != nil {
		log.Fatal(err)
	}
}
