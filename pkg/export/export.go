package export

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/db"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/gh"
)

func ExportAndSaveJobs(ctx context.Context, conf config.ConfigType, runAttempt int64, exitOnTokenRateLimit bool) error {
	jobsInfo, rate, err := gh.GetWorkflowAttemptJobs(ctx, conf, runAttempt)
	if err != nil {
		log.Fatal(err)
	}
	if rate.Remaining < 20 {
		fmt.Printf("Close to rate limit, remaining: %d", rate.Remaining)
		if exitOnTokenRateLimit {
			fmt.Printf("Exit due to the flag -exit-on-token-rate-limit=true")
			return nil
		}
		fmt.Printf("Sleep till %v (%v seconds)\n", rate.Reset, time.Until(rate.Reset.Time))
		time.Sleep(time.Until(rate.Reset.Time))
	}
	var dbContext config.DbContextType
	db.PrepareJobTransaction(ctx, conf, &dbContext)
	for _, jobInfo := range jobsInfo {
		err = db.SaveJobInfo(&dbContext, jobInfo)
		if err != nil {
			fmt.Println(err)
		}
	}
	err = db.CommitJobTransaction(&dbContext)

	return err
}
