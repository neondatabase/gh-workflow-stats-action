package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/neondatabase/gh-workflow-stats-action/internal/db"
	"github.com/neondatabase/gh-workflow-stats-action/internal/exporter"
	"github.com/neondatabase/gh-workflow-stats-action/internal/githubclient"
)

func NewExportCommand() *cobra.Command {
	dbCfg := &db.Config{}
	ghCfg := &githubclient.Config{}
	ghRunID := int64(0)
	ghRunAttempt := int64(0)

	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: buildRunCommand(dbCfg, ghCfg, ghRunID, ghRunAttempt),
	}

	exportCmd.Flags().StringVar(&dbCfg.DSN, "db-uri", "", "DB DSN")
	exportCmd.Flags().StringVar(&dbCfg.TableName, "db-table", "", "DB Table")
	exportCmd.Flags().StringVar(&ghCfg.Repository, "github-repository", "", "Github repository")
	exportCmd.Flags().StringVar(&ghCfg.Token, "github-token", "", "Github token")
	exportCmd.Flags().Int64Var(&ghRunID, "github-run-id", 0, "Github run ID")
	exportCmd.Flags().Int64Var(&ghRunAttempt, "github-run-attempt", 0, "Github run ID")

	exportCmd.MarkFlagsRequiredTogether("github-repository", "github-run-id", "github-token", "github-run-attempt")

	exportCmd.MarkFlagRequired("db-uri")
	exportCmd.MarkFlagRequired("github-repository")

	return exportCmd
}

func init() {
	rootCmd.AddCommand(NewExportCommand())
}

func buildRunCommand(dbCfg *db.Config, ghCfg *githubclient.Config, ghRunID int64, ghRunAttempt int64) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		repo, err := db.NewDatabase(dbCfg)
		if err != nil {
			fmt.Println(err)
			return
		}

		ghClient, err := githubclient.NewClient(ghCfg)
		if err != nil {
			fmt.Println(err)
			return
		}

		exp := exporter.New(repo, ghClient)
		if err := exp.ExportByRunAttempt(context.Background(), ghRunID, ghRunAttempt); err != nil {
			fmt.Println(err)
			return
		}
	}
}
