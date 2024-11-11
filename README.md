# Github Workflow Stats exporter to Postgres

## WIP
Work is in progress so list of parameters could be changed.

## Inputs

| Input               | Description                               |
| ------------------- | ----------------------------------------- |
| `db_uri`            | Database URI                              |
| `db_table`          | Table for storing Workflow stats          |
| `gh_run_id`         | Workflow Run Id to get information on     |
| `gh_token`          | Github Token, optional for public Repo    |
| `duration`          | Duration for the history period to export |
