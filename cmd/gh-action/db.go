package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	schemeWorkflowRunsStats = `
	CREATE TABLE %s (
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
	schemeWorkflowRuns = `
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
	db, err := sqlx.Connect("postgres", conf.dbUri)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(schemeWorkflowRunsStats, conf.dbTable))
	if err != nil {
		return err
	}
	return nil
}

func saveRecords(conf configType, records *WorkflowStat) error {
	db, err := sqlx.Connect("postgres", conf.dbUri)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", conf.dbTable,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err = db.NamedExec(query, *records)

	if err != nil {
		return err
	}
	return nil
}
