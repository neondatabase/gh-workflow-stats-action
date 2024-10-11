package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/db"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/gh"
)

func main() {
	var dateStr string
	var date time.Time

	flag.StringVar(&dateStr, "date", "", "date to quert and export")
	flag.Parse()

	if dateStr == "" {
		date = time.Now().Truncate(24 * time.Hour)
	}else {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatalf("Failed to parse date: %s", err)
		}
	}

	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = db.ConnectDB(&conf)
	if err != nil {
		log.Fatal(err)
	}

	gh.InitGhClient(&conf)
	ctx := context.Background()

	endDate := date.Add(1 * time.Hour)
	for date.Before(endDate) {
		runs, _ := gh.ListWorkflowRuns(ctx, conf, date, date.Add(16*time.Hour))
		fmt.Println(date, len(runs), date.Format(time.RFC3339))
		for _, rec := range(runs) {
			fmt.Printf("%s-%d-%d, ", rec.GetName(), rec.GetRunAttempt(), rec.GetID())
		}
		fmt.Println()
		date = date.Add(time.Hour)
	}
}
