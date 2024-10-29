package exporter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/jmoiron/sqlx"

	"github.com/neondatabase/gh-workflow-stats-action/internal/db"
	"github.com/neondatabase/gh-workflow-stats-action/internal/githubclient"
)

type Export struct {
	repo     db.Repository
	ghClient githubclient.Client
}

func New(repo db.Repository, ghClient githubclient.Client) *Export {
	return &Export{
		repo:     repo,
		ghClient: ghClient,
	}
}

func (e *Export) ExportByInterval(ctx context.Context, startDate, endDate time.Time) error {
	durations := []time.Duration{
		6 * time.Hour, // 18:00 - 24:00
		3 * time.Hour, // 15:00 - 18:00
		1 * time.Hour, // 14:00 - 15:00
		1 * time.Hour, // 13:00 - 14:00
		1 * time.Hour, // 12:00 - 13:00
		2 * time.Hour, // 10:00 - 12:00
		4 * time.Hour, // 06:00 - 10:00
		6 * time.Hour, // 00:00 - 06:00
	}
	curDurIdx := 0
	for date := endDate.Add(-durations[curDurIdx]); date.Compare(startDate) >= 0; date = date.Add(-durations[curDurIdx]) {
		runs, rate, _ := e.ghClient.ListWorkflowRuns(ctx, date, date.Add(durations[curDurIdx]))
		fmt.Println("\n", date, len(runs))
		if len(runs) >= 1000 {
			fmt.Printf("\n\n+++\n+ PAGINATION LIMIT: %v\n+++\n", date)
		}
		fetchedRunsKeys := make([]githubclient.WorkflowRunAttemptKey, len(runs))
		for key := range runs {
			fetchedRunsKeys = append(fetchedRunsKeys, key)
		}
		notInDb := e.repo.QueryWorkflowRunsNotInDb(fetchedRunsKeys)
		fmt.Printf("Time range: %v - %v, fetched: %d, notInDb: %d.\n",
			date, date.Add(durations[curDurIdx]),
			len(runs), len(notInDb),
		)
		if rate.Remaining < 30 {
			fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
			fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
			time.Sleep(time.Until(rate.Reset.Time) + 10*time.Second)
		} else {
			fmt.Printf("Rate: %+v\n", rate)
		}
		for _, key := range notInDb {
			fmt.Printf("Saving runId %d Attempt %d. ", key.RunId, key.RunAttempt)
			var attemptRun *github.WorkflowRun
			var ok bool
			if attemptRun, ok = runs[githubclient.WorkflowRunAttemptKey{RunId: key.RunId, RunAttempt: key.RunAttempt}]; ok {
				fmt.Printf("Got it from ListWorkflowRuns results. ")
			} else {
				fmt.Printf("Fetching it from GH API. ")
				attemptRun, _ = e.ghClient.GetWorkflowAttempt(ctx, key.RunId, key.RunAttempt)
			}
			e.repo.SaveWorkflowRunAttempt(attemptRun)
			e.ExportByRunAttempt(ctx, key.RunId, key.RunAttempt)
		}
		curDurIdx = (curDurIdx + 1) % len(durations)
	}

	return nil
}

func (e *Export) ExportByRunAttempt(ctx context.Context, runID int64, runAttempt int64) error {
	jobsInfo, rate, err := e.ghClient.GetWorkflowAttemptJobs(ctx, runID, runAttempt)
	if err != nil {
		log.Fatal(err)
	}
	if rate.Remaining < 20 {
		fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
		fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
		time.Sleep(time.Until(rate.Reset.Time))
	}
	err = e.repo.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		for _, jobInfo := range jobsInfo {
			err = e.repo.InsertJob(tx, jobInfo)
			if err != nil {
				return err
			}

			err = e.repo.InsertSteps(tx, jobInfo)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
