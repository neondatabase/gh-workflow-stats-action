package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/jmoiron/sqlx"
)

type ConfigType struct {
	DbUri       string
	DbTable     string
	Db          *sqlx.DB
	RunID       int64
	Repository  string
	Owner       string
	Repo        string
	GithubToken string
	GhClient    *github.Client
}

type DbContextType struct {
	Tx             *sqlx.Tx
	InsertJobStmt  *sqlx.NamedStmt
	InsertStepStmt *sqlx.NamedStmt
}

func GetConfig() (ConfigType, error) {
	dbUri := os.Getenv("DB_URI")
	if len(dbUri) == 0 {
		return ConfigType{}, fmt.Errorf("missing env: DB_URI")
	}

	dbTable := os.Getenv(("DB_TABLE"))
	if len(dbTable) == 0 {
		return ConfigType{}, fmt.Errorf("missing env: DB_TABLE")
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if len(repository) == 0 {
		return ConfigType{}, fmt.Errorf("missing env: GITHUB_REPOSITORY")
	}

	envRunID := os.Getenv("GH_RUN_ID")
	var runID int64
	if len(envRunID) == 0 {
		return ConfigType{}, fmt.Errorf("missing env: GH_RUN_ID")
	}
	runID, err := strconv.ParseInt(envRunID, 10, 64)
	if err != nil {
		return ConfigType{}, fmt.Errorf("GH_RUN_ID must be integer, error: %v", err)
	}

	githubToken := os.Getenv("GH_TOKEN")

	repoDetails := strings.Split(repository, "/")
	if len(repoDetails) != 2 {
		return ConfigType{}, fmt.Errorf("invalid env: GITHUB_REPOSITORY")
	}

	return ConfigType{
		DbUri:       dbUri,
		DbTable:     dbTable,
		Db:          nil,
		RunID:       runID,
		Repository:  repository,
		Owner:       repoDetails[0],
		Repo:        repoDetails[1],
		GithubToken: githubToken,
	}, nil
}
