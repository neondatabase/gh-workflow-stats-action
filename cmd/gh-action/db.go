package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/google/go-github/v65/github"
)

var (
	schemeWorkflowRunsStats = `
	CREATE TABLE IF NOT EXISTS %s (
		workflowid BIGINT,
		name	   TEXT,
		status     TEXT,
		conclusion TEXT,
		runid      INT,
		runattempt INT,
		startedat  TIMESTAMP,
		updatedat  TIMESTAMP,
		reponame   TEXT,
		event      TEXT,
		PRIMARY KEY(workflowid, runattempt)
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
		JobId		BIGINT
		RunID		BIGINT
		NodeID 		TEXT
		HeadBranch	TEXT
		HeadSHA		TEXT
		Status		TEXT
		Conclusion	TEXT
		CreatedAt	TIMESTAMP
		StartedAt	TIMESTAMP
		CompletedAt	TIMESTAMP
		Name		TEXT
		RunnerName	TEXT
		RunnerGroupName	TEXT
		RunAttempt		BIGINT
		WorkflowName	TEXT
	)`
	schemeWorkflowJobsSteps = `
	CREATE TABLE IF NOT EXISTS %s (
		JobId		BIGINT
		RunId		BIGINT
		Name		TEXT
		Status		TEXT
		Conclusion	TEXT
		Number		BIGINT
		StartedAt	TIMESTAMP
		CompletedAt	TIMESTAMP
	)
	`
)

func initDatabase(conf configType) error {
	_, err := conf.db.Exec(fmt.Sprintf(schemeWorkflowRunsStats, conf.dbTable))
	if err != nil {
		return err
	}

	_, err = conf.db.Exec(fmt.Sprintf(schemeWorkflowRunAttempts, conf.dbTable + "_attempts"))
	if err != nil {
		return err
	}

	_, err = conf.db.Exec(fmt.Sprintf(schemeWorkflowJobs, conf.dbTable + "_jobs"))
	if err != nil {
		return err
	}

	_, err = conf.db.Exec(fmt.Sprintf(schemeWorkflowJobsSteps, conf.dbTable + "_steps"))
	if err != nil {
		return err
	}
	return nil
}

func connectDB(conf *configType) error {
	db, err := sqlx.Connect("postgres", conf.dbUri)
	if err != nil {
		return err
	}
	conf.db = db
	return nil
}

func saveWorkflowRun(conf configType, record *WorkflowRunRec) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", conf.dbTable,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := conf.db.NamedExec(query, *record)

	if err != nil {
		return err
	}
	return nil
}

func saveWorkflowRunAttempt(conf configType, workflowRun *github.WorkflowRun) error {
	query := fmt.Sprintf("INSERT INTO %s_attempts (%s) VALUES (%s)", conf.dbTable,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := conf.db.NamedExec(query, ghWorkflowRunRec(workflowRun))
	return err
}
