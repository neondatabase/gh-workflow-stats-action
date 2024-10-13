package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v65/github"
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

	durations := []time.Duration{
		6 * time.Hour,	// 18:00 - 24:00
		3 * time.Hour,  // 15:00 - 18:00
		1 * time.Hour,  // 14:00 - 15:00
		1 * time.Hour,  // 13:00 - 14:00
		1 * time.Hour,  // 12:00 - 13:00
		2 * time.Hour,  // 10:00 - 12:00
		4 * time.Hour,	// 06:00 - 10:00
		6 * time.Hour,	// 00:00 - 06:00
	}
	curDurIdx := 0
	for date := endDate.Add(-durations[curDurIdx]); date.Compare(startDate) >= 0; date = date.Add(-durations[curDurIdx]) {
		runs, rate, _ := gh.ListWorkflowRuns(ctx, conf, date, date.Add(durations[curDurIdx]))
		fmt.Println("\n", date, len(runs))
		if len(runs) >= 1000 {
			fmt.Printf("\n\n+++\n+ PAGINATION LIMIT: %v\n+++\n", date)
		}
		if rate.Remaining < 30 {
			fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
			fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
			time.Sleep(time.Until(rate.Reset.Time) + 10*time.Second)
		}else {
			fmt.Printf("Rate: %+v\n", rate)
		}
		runIdSet := make(map[int64]struct{})
		for key, rec := range(runs) {
			conf.RunID = key.RunId
			storedAttempts := db.QueryWorkflowRunAttempts(conf, key.RunId)
			var attempt int64
			for attempt = 1; attempt < int64(rec.GetRunAttempt())+1; attempt++ {
				if _, ok := storedAttempts[attempt]; ok {
					fmt.Printf("\nRunId %d Attempt %d already in database, skip. ", rec.GetID(), attempt)
				}else {
					fmt.Printf("Saving runId %d Attempt %d. ", rec.GetID(), attempt)
					var attemptRun *github.WorkflowRun
					if attemptRun, ok = runs[gh.WorkflowRunAttemptKey{RunId: key.RunId, Attempt: attempt}]; ok {
						fmt.Printf("Got it from ListWorkflowRuns results. ")
					}else {
						fmt.Printf("Fetching it from GH API. ")
						attemptRun, _ = gh.GetWorkflowAttempt(ctx, conf, attempt)
					}
					db.SaveWorkflowRunAttempt(conf, attemptRun)
					export.ExportAndSaveJobs(ctx, conf, attempt)
				}
			}
			runIdSet[rec.GetID()] = struct{}{}
		}
		curDurIdx = (curDurIdx + 1) % len(durations)
	}
}
