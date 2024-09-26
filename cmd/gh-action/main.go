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

	var records *WorkflowStat
	records, err = createRecords(ctx, conf)
	if err != nil {
		log.Fatal(err)
	}

	err = saveRecords(conf, records)
	if err != nil {
		log.Fatal(err)
	}
}
