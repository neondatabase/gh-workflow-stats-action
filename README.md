# Github Workflow Stats exporter to Postgres

## Inputs

| Input               | Description                            |
| ------------------- | -------------------------------------- |
| `DB_URI`            | Database URI                           |
| `DB_TABLE`          | Table for storing Workflow stats       |
| `GH_RUN_ID`         | Workflow Run Id to get information on  |
| `GH_TOKEN`          | Github Token, optional for public Repo |
