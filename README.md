# Creating the table
In this example, I'll be using the wonderful [usql](https://github.com/xubingnan123/usql) client.

## Connect to the database:
To connect, run the following, replacing the variables with their corresponding values:

`usql postgres://${COMMANDS_DB_USER}@${COMMANDS_DB_HOST}:${COMMANDS_DB_PORT}/${COMMANDS_DB_NAME}`

You should then be at a SQL prompt that looks something like the following:

`pg:commands@commands-db/logging=>`

## Create logging table
To create a table with the proper structure, run the following (as always, adjusting variables as needed):
```
CREATE TABLE ${COMMANDS_DB_TABLE} (
	id SERIAL PRIMARY KEY,
	starttime timestamp NOT NULL,
	stoptime timestamp NOT NULL,
	hostname varchar NOT NULL,
	commandname varchar NOT NULL,
	exitcode int NOT NULL
);
```

## Configure the container
The following environment variables are used to configure the service (all values provided are just examples):
```
COMMANDS_DB_TYPE=postgresql
COMMANDS_DB_HOST=commands-db
COMMANDS_DB_PORT=5432
COMMANDS_DB_USER=commands
COMMANDS_DB_PASS=changeme
COMMANDS_DB_NAME=logging
COMMANDS_DB_TABLE=logging
COMMANDS_DB_SSL_MODE=disable
COMMANDS_PORT=8080
TZ=America/Chicago
```
