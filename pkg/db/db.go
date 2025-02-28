package db

// Trigger

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/google/go-github/v65/github"

	"github.com/neondatabase/gh-workflow-stats-action/pkg/config"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/gh"
)

var (
	schemeWorkflowRunsStats = `
	CREATE TABLE IF NOT EXISTS %s (
		workflowid BIGINT,
		name	   TEXT,
		status     TEXT,
		conclusion TEXT,
		runid      BIGINT,
		runattempt INT,
		startedat  TIMESTAMP,
		updatedat  TIMESTAMP,
		reponame   TEXT,
		event      TEXT,
		PRIMARY KEY(workflowid, runid, runattempt)
	)
	`
	schemeWorkflowRunAttempts = `
	CREATE TABLE IF NOT EXISTS %s (
		workflowid BIGINT,
		name	   TEXT,
		status     TEXT,
		conclusion TEXT,
		runid      BIGINT,
		runattempt INT,
		startedat  TIMESTAMP,
		updatedat  TIMESTAMP,
		reponame   TEXT,
		event      TEXT,
		PRIMARY KEY(workflowid, runid, runattempt)
	)
	`
	schemeWorkflowJobs = `
	CREATE TABLE IF NOT EXISTS %s (
		JobId		BIGINT,
		RunID		BIGINT,
		NodeID 		TEXT,
		HeadBranch	TEXT,
		HeadSHA		TEXT,
		Status		TEXT,
		Conclusion	TEXT,
		CreatedAt	TIMESTAMP,
		StartedAt	TIMESTAMP,
		CompletedAt	TIMESTAMP,
		Name		TEXT,
		RunnerName	TEXT,
		RunnerGroupName	TEXT,
		RunAttempt		BIGINT,
		WorkflowName	TEXT
	)`
	schemeWorkflowJobsSteps = `
	CREATE TABLE IF NOT EXISTS %s (
		JobId		BIGINT,
		RunId		BIGINT,
		RunAttempt	BIGINT,
		Name		TEXT,
		Status		TEXT,
		Conclusion	TEXT,
		Number		BIGINT,
		StartedAt	TIMESTAMP,
		CompletedAt	TIMESTAMP
	)
	`
)

func InitDatabase(conf config.ConfigType) error {
	_, err := conf.Db.Exec(fmt.Sprintf(schemeWorkflowRunsStats, conf.DbTable))
	if err != nil {
		return err
	}

	_, err = conf.Db.Exec(fmt.Sprintf(schemeWorkflowRunAttempts, conf.DbTable+"_attempts"))
	if err != nil {
		return err
	}

	_, err = conf.Db.Exec(fmt.Sprintf(schemeWorkflowJobs, conf.DbTable+"_jobs"))
	if err != nil {
		return err
	}

	_, err = conf.Db.Exec(fmt.Sprintf(schemeWorkflowJobsSteps, conf.DbTable+"_steps"))
	if err != nil {
		return err
	}
	return nil
}

func ConnectDB(conf *config.ConfigType) error {
	db, err := sqlx.Connect("postgres", conf.DbUri)
	if err != nil {
		return err
	}
	conf.Db = db
	return nil
}

func SaveWorkflowRun(conf config.ConfigType, record *data.WorkflowRunRec) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", conf.DbTable,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := conf.Db.NamedExec(query, *record)

	if err != nil {
		return err
	}
	return nil
}

func SaveWorkflowRunAttempt(conf config.ConfigType, workflowRun *github.WorkflowRun) error {
	query := fmt.Sprintf("INSERT INTO %s_attempts (%s) VALUES (%s)", conf.DbTable,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := conf.Db.NamedExec(query, data.GhWorkflowRunRec(workflowRun))
	return err
}

func PrepareJobTransaction(ctx context.Context, conf config.ConfigType, dbContext *config.DbContextType) error {
	var err error
	dbContext.Tx, err = conf.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	jobs_query := fmt.Sprintf("INSERT INTO %s_jobs (%s) VALUES (%s)", conf.DbTable,
		"jobid, runid, nodeid, headbranch, headsha, status, conclusion, createdat, startedat, completedat, name, runnername, runnergroupname, runattempt, workflowname",
		":jobid, :runid, :nodeid, :headbranch, :headsha, :status, :conclusion, :createdat, :startedat, :completedat, :name, :runnername, :runnergroupname, :runattempt, :workflowname",
	)
	dbContext.InsertJobStmt, _ = dbContext.Tx.PrepareNamed(jobs_query)

	steps_query := fmt.Sprintf("INSERT INTO %s_steps (%s) VALUES (%s)", conf.DbTable,
		"jobid, runid, runattempt, name, status, conclusion, number, startedat, completedat",
		":jobid, :runid, :runattempt, :name, :status, :conclusion, :number, :startedat, :completedat",
	)
	dbContext.InsertStepStmt, _ = dbContext.Tx.PrepareNamed(steps_query)

	return nil
}
func SaveJobInfo(dbContext *config.DbContextType, workflowJob *github.WorkflowJob) error {
	_, err := dbContext.InsertJobStmt.Exec(data.GhWorkflowJobRec(workflowJob))
	if err != nil {
		return err
	}

	for _, step := range workflowJob.Steps {
		err = SaveStepInfo(dbContext, workflowJob, step)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveStepInfo(dbContext *config.DbContextType, job *github.WorkflowJob, step *github.TaskStep) error {
	_, err := dbContext.InsertStepStmt.Exec(data.GhWorkflowJobStepRec(job, step))
	return err
}

func CommitJobTransaction(dbContext *config.DbContextType) error {
	err := dbContext.Tx.Commit()
	return err
}

func QueryWorkflowRunAttempts(conf config.ConfigType, runId int64) map[int64]struct{} {
	result := make(map[int64]struct{})

	query := fmt.Sprintf("SELECT runAttempt from %s_attempts WHERE runId=$1", conf.DbTable)
	rows, err := conf.Db.Query(query, runId)
	if err != nil {
		return result
	}
	var attempt int64
	for rows.Next() {
		err = rows.Scan(&attempt)
		if err != nil {
			fmt.Println(err)
		} else {
			result[attempt] = struct{}{}
		}
	}
	return result
}

func QueryWorkflowRunsNotInDb(conf config.ConfigType, workflowRuns []gh.WorkflowRunAttemptKey) []gh.WorkflowRunAttemptKey {
	result := make([]gh.WorkflowRunAttemptKey, 0)

	if len(workflowRuns) == 0 {
		return result
	}
	// TODO: I have to find out how to use https://jmoiron.github.io/sqlx/#namedParams with sqlx.In()
	// For now just generate query with strings.Builder
	var valuesStr strings.Builder
	for i, v := range workflowRuns {
		if i > 0 {
			valuesStr.WriteString(", ")
		}
		valuesStr.WriteString(fmt.Sprintf("(%d :: bigint, %d :: bigint)", v.RunId, v.RunAttempt))
	}
	queryStr := fmt.Sprintf("SELECT runid, runattempt FROM (VALUES %s) as q (runid, runattempt) LEFT JOIN %s_attempts db "+
		"USING (runid, runattempt) WHERE db.runid is null",
		valuesStr.String(),
		conf.DbTable,
	)
	rows, err := conf.Db.Queryx(queryStr)
	if err != nil {
		fmt.Printf("Failed to Query: %s\n", err)
		return result
	}
	var rec gh.WorkflowRunAttemptKey
	for rows.Next() {
		err = rows.StructScan(&rec)
		if err != nil {
			fmt.Println(err)
		} else {
			result = append(result, rec)
		}
	}
	return result
}
