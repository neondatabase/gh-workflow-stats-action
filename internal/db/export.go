package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/neondatabase/gh-workflow-stats-action/internal/githubclient"
	"github.com/neondatabase/gh-workflow-stats-action/pkg/data"
)

var _ Repository = &Postgres{}

type Postgres struct {
	cfg  *Config
	conn *sqlx.DB
}

func NewDatabase(cfg *Config) (*Postgres, error) {
	conn, err := sqlx.Connect("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		cfg:  cfg,
		conn: conn,
	}, nil
}

func (p *Postgres) Init() error {
	_, err := p.conn.Exec(fmt.Sprintf(schemeWorkflowRunsStats, p.cfg.TableName))
	if err != nil {
		return err
	}

	_, err = p.conn.Exec(fmt.Sprintf(schemeWorkflowRunAttempts, p.cfg.TableName+"_attempts"))
	if err != nil {
		return err
	}

	_, err = p.conn.Exec(fmt.Sprintf(schemeWorkflowJobs, p.cfg.TableName+"_jobs"))
	if err != nil {
		return err
	}

	_, err = p.conn.Exec(fmt.Sprintf(schemeWorkflowJobsSteps, p.cfg.TableName+"_steps"))
	if err != nil {
		return err
	}
	return nil
}

func (p *Postgres) SaveWorkflowRun(record *data.WorkflowRunRec) error {
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", p.cfg.TableName,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := p.conn.NamedExec(query, *record)

	if err != nil {
		return err
	}
	return nil
}

func (p *Postgres) SaveWorkflowRunAttempt(workflowRun *github.WorkflowRun) error {
	query := fmt.Sprintf("INSERT INTO %s_attempts (%s) VALUES (%s)", p.cfg.TableName,
		"workflowid, name, status, conclusion, runid, runattempt, startedAt, updatedAt, repoName, event",
		":workflowid, :name, :status, :conclusion, :runid, :runattempt, :startedat, :updatedat, :reponame, :event",
	)

	_, err := p.conn.NamedExec(query, data.GhWorkflowRunRec(workflowRun))
	return err
}

func (p *Postgres) WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := p.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err = fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			fmt.Println(err)
		}

		return err
	}

	return tx.Commit()
}

func (p *Postgres) InsertJob(tx *sqlx.Tx, workflowJob *github.WorkflowJob) error {
	query := fmt.Sprintf("INSERT INTO %s_jobs (%s) VALUES (%s)", p.cfg.TableName,
		"jobid, runid, nodeid, headbranch, headsha, status, conclusion, createdat, startedat, completedat, name, runnername, runnergroupname, runattempt, workflowname",
		":jobid, :runid, :nodeid, :headbranch, :headsha, :status, :conclusion, :createdat, :startedat, :completedat, :name, :runnername, :runnergroupname, :runattempt, :workflowname",
	)
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(data.GhWorkflowJobRec(workflowJob))
	if err != nil {
		return err
	}

	return nil
}

func (p *Postgres) InsertSteps(tx *sqlx.Tx, job *github.WorkflowJob) error {
	query := fmt.Sprintf("INSERT INTO %s_steps (%s) VALUES (%s)", p.cfg.TableName,
		"jobid, runid, runattempt, name, status, conclusion, number, startedat, completedat",
		":jobid, :runid, :runattempt, :name, :status, :conclusion, :number, :startedat, :completedat",
	)
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		return err
	}

	for _, step := range job.Steps {
		_, err = stmt.Exec(data.GhWorkflowJobStepRec(job, step))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Postgres) QueryWorkflowRunAttempts(runId int64) map[int64]struct{} {
	result := make(map[int64]struct{})

	query := fmt.Sprintf("SELECT runAttempt from %s_attempts WHERE runId=$1", p.cfg.TableName)
	rows, err := p.conn.Query(query, runId)
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

func (p *Postgres) QueryWorkflowRunsNotInDb(workflowRuns []githubclient.WorkflowRunAttemptKey) []githubclient.WorkflowRunAttemptKey {
	result := make([]githubclient.WorkflowRunAttemptKey, 0)

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
		p.cfg.TableName,
	)
	rows, err := p.conn.Queryx(queryStr)
	if err != nil {
		fmt.Printf("Failed to Query: %s\n", err)
		return result
	}
	var rec githubclient.WorkflowRunAttemptKey
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
