package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v65/github"
	"github.com/jmoiron/sqlx"
)

type configType struct {
	dbUri       string
	dbTable     string
	db			*sqlx.DB
	runID       int64
	repository  string
	owner       string
	repo        string
	githubToken string
	ghClient    *github.Client
}

func getConfig() (configType, error) {
	dbUri := os.Getenv("DB_URI")
	if len(dbUri) == 0 {
		return configType{}, fmt.Errorf("missing env: DB_URI")
	}

	dbTable := os.Getenv(("DB_TABLE"))
	if len(dbTable) == 0 {
		return configType{}, fmt.Errorf("missing env: DB_TABLE")
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if len(repository) == 0 {
		return configType{}, fmt.Errorf("missing env: GITHUB_REPOSITORY")
	}

	envRunID := os.Getenv("GH_RUN_ID")
	var runID int64
	if len(envRunID) == 0 {
		return configType{}, fmt.Errorf("missing env: GH_RUN_ID")
	}
	runID, err := strconv.ParseInt(envRunID, 10, 64)
	if err != nil {
		return configType{}, fmt.Errorf("GH_RUN_ID must be integer, error: %v", err)
	}

	githubToken := os.Getenv("GH_TOKEN")

	repoDetails := strings.Split(repository, "/")
	if len(repoDetails) != 2 {
		return configType{}, fmt.Errorf("invalid env: GITHUB_REPOSITORY")
	}

	return configType{
		dbUri:       dbUri,
		dbTable:     dbTable,
		db:          nil,
		runID:       runID,
		repository:  repository,
		owner:       repoDetails[0],
		repo:        repoDetails[1],
		githubToken: githubToken,
	}, nil
}
