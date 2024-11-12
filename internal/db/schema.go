package db

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
