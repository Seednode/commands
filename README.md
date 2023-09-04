## Creating the table
In this example, I'll be using the wonderful [usql](https://github.com/xubingnan123/usql) client.

# Connect to the database:
To connect, run the following, adjusting variables as needed:

`usql postgres://${COMMANDS_DB_USER}@${COMMANDS_DB_HOST}:${COMMANDS_DB_PORT}/${COMMANDS_DB_NAME}`

# Create logging table
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

# Configure the container
The following environment variables are used to configure the service:
```
COMMANDS_DB_TYPE=
COMMANDS_DB_HOST=
COMMANDS_DB_PORT=
COMMANDS_DB_USER=
COMMANDS_DB_PASS=
COMMANDS_DB_NAME=
COMMANDS_DB_TABLE=
COMMANDS_DB_SSL_MODE=
COMMANDS_PORT=
```
