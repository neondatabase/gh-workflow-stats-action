package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/db"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/export"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/gh"
)

const (
	queryPeriod = 2 * time.Hour
)

func main() {
	var startDateStr string
	var endDateStr string
	var startDate time.Time
	var endDate time.Time

	flag.StringVar(&startDateStr, "start-date", "", "start date to quert and export")
	flag.StringVar(&endDateStr, "end-date", "", "end date to quert and export")
	flag.Parse()

	if startDateStr == "" {
		startDate = time.Now().Truncate(24 * time.Hour)
	}else {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			log.Fatalf("Failed to parse date: %s", err)
		}
	}

	if endDateStr == "" {
		endDate = startDate.AddDate(0, 0, 1)
	}else {
		var err error
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			log.Fatalf("Failed to parse end date: %s", err)
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

	for date := endDate; date.After(startDate); date = date.Add(-queryPeriod) {
		runs, rate, _ := gh.ListWorkflowRuns(ctx, conf, date, date.Add(queryPeriod))
		fmt.Println("\n", date, len(runs))
		if len(runs) >= 1000 {
			fmt.Printf("\n\n+++\n+ PAGINATION LIMIT: %v\n+++\n", date)
		}
		if rate.Remaining < 200 {
			fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
			fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
			time.Sleep(time.Until(rate.Reset.Time) + 10*time.Second)
		}else {
			fmt.Printf("Rate: %+v\n", rate)
		}
		runIdSet := make(map[int64]struct{})
		for _, rec := range(runs) {
			conf.RunID = rec.GetID()
			storedAttempts := db.QueryWorkflowRunAttempts(conf, rec.GetID())
			var attempt int64
			for attempt = 1; attempt < int64(rec.GetRunAttempt())+1; attempt++ {
				if _, ok := storedAttempts[attempt]; ok {
					fmt.Printf("\nRunId %d Attempt %d already in database, skip. ", rec.GetID(), attempt)
				}else {
					fmt.Printf("Saving runId %d Attempt %d.", rec.GetID(), attempt)
					attemptRun, _ := gh.GetWorkflowAttempt(ctx, conf, attempt)
					db.SaveWorkflowRunAttempt(conf, attemptRun)
					export.ExportAndSaveJobs(ctx, conf, attempt)
				}
			}
			runIdSet[rec.GetID()] = struct{}{}
		}
	}
}
