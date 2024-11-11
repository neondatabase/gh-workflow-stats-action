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

func parseTimesAndDuration(startTimeStr string, endTimeStr string, durationStr string) (time.Time, time.Time, error) {
	startTime := time.Now()
	endTime := time.Now()
	checkTimeLayouts := []string{time.DateOnly, time.RFC3339, time.DateTime}
	var duration time.Duration
	var err error

	if startTimeStr != "" && endTimeStr != "" && durationStr != "" {
		return startTime, endTime, fmt.Errorf("you can't set all startTime, endTime and duration")
	}

	if durationStr != "" {
		duration, err = time.ParseDuration(durationStr)
		if err != nil {
			return startTime, endTime, fmt.Errorf("failed to parse duration [%s]: %v", durationStr, err)
		}
	}
	if endTimeStr != "" {
		parsedSuccess := false
		for _, layout := range checkTimeLayouts {
			endTime, err = time.Parse(layout, endTimeStr)
			if err == nil {
				parsedSuccess = true
				break
			}
		}
		if !parsedSuccess {
			return startTime, endTime, fmt.Errorf("failed to parse endTime [%s]: %v", endTimeStr, err)
		}
	}
	if startTimeStr != "" {
		parsedSuccess := false
		for _, layout := range checkTimeLayouts {
			startTime, err = time.Parse(layout, startTimeStr)
			if err == nil {
				parsedSuccess = true
				break
			}
		}
		if !parsedSuccess {
			return startTime, endTime, fmt.Errorf("failed to parse startTime [%s]: %v", startTimeStr, err)
		}
	}
	if startTimeStr == "" && endTimeStr == "" {
		startTime = time.Now().Truncate(24 * time.Hour)
		endTime = time.Now().Truncate(time.Minute)
	}

	if durationStr != "" {
		if startTimeStr != "" {
			return startTime, startTime.Add(duration), nil
		}
		return endTime.Add(-duration), endTime, nil
	}

	return startTime, endTime, nil
}

func main() {
	var startTimeStr string
	var endTimeStr string
	var durationStr string
	var startTime time.Time
	var endTime time.Time
	var err error

	var exitOnTokenRateLimit bool

	flag.StringVar(&startTimeStr, "start-time", "", "start time to query and export")
	flag.StringVar(&endTimeStr, "end-time", "", "end time to query and export")
	flag.StringVar(&durationStr, "duration", "", "duration of the export period")
	flag.BoolVar(&exitOnTokenRateLimit, "exit-on-token-rate-limit", false, "Should program exit when we hit github token rate limit or sleep and wait for renewal")
	flag.Parse()

	startTime, endTime, err = parseTimesAndDuration(startTimeStr, endTimeStr, durationStr)
	if err != nil {
		log.Fatalf("Failed to parse dates: %s", err)
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

	queryDuration := time.Duration(time.Hour)
	for queryTime := endTime.Add(-queryDuration); queryTime.Compare(startTime) >= 0; queryTime = queryTime.Add(-queryDuration) {
		runs, rate, _ := gh.ListWorkflowRuns(ctx, conf, queryTime, queryTime.Add(queryDuration))
		fmt.Println("\n", queryTime, len(runs))
		if len(runs) >= 1000 {
			fmt.Printf("\n\n+++\n+ PAGINATION LIMIT: %v\n+++\n", queryTime)
		}
		fetchedRunsKeys := make([]gh.WorkflowRunAttemptKey, len(runs))
		i := 0
		for key := range runs {
			fetchedRunsKeys[i] = key
			i++
		}
		notInDb := db.QueryWorkflowRunsNotInDb(conf, fetchedRunsKeys)
		fmt.Printf("Time range: %v - %v, fetched: %d, notInDb: %d.\n",
			queryTime, queryTime.Add(queryDuration),
			len(runs), len(notInDb),
		)
		if rate.Remaining < 30 {
			if exitOnTokenRateLimit {
				break
			}
			fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
			fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
			time.Sleep(time.Until(rate.Reset.Time))
		} else {
			fmt.Printf("Rate: %+v\n", rate)
		}
		for _, key := range notInDb {
			conf.RunID = key.RunId
			fmt.Printf("Saving runId %d Attempt %d. ", key.RunId, key.RunAttempt)
			var attemptRun *github.WorkflowRun
			var ok bool
			if attemptRun, ok = runs[gh.WorkflowRunAttemptKey{RunId: key.RunId, RunAttempt: key.RunAttempt}]; ok {
				fmt.Printf("Got it from ListWorkflowRuns results. ")
			} else {
				fmt.Printf("Fetching it from GH API. ")
				attemptRun, _ = gh.GetWorkflowAttempt(ctx, conf, key.RunAttempt)
			}
			db.SaveWorkflowRunAttempt(conf, attemptRun)
			export.ExportAndSaveJobs(ctx, conf, key.RunAttempt, exitOnTokenRateLimit)
		}
	}
}
