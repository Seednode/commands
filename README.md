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

## Usage output
Alternatively, you can configure the service using command-line flags.
```
Display command logs from a database.

Usage:
  commands [flags]

Flags:
  -b, --bind string                 address to bind to (default "0.0.0.0")
      --database-host string        database host to connect to
      --database-name string        database name to connect to
      --database-pass string        database password to connect with
      --database-port string        database port to connect to
      --database-root-cert string   database ssl root certificate path
      --database-ssl-cert string    database ssl connection certificate path
      --database-ssl-key string     database ssl connection key path
      --database-ssl-mode string    database ssl connection mode
      --database-table string       database table to query
      --database-type string        database type to connect to
      --database-user string        database user to connect as
  -h, --help                        help for commands
  -p, --port uint16                 port to listen on (default 8080)
      --time-zone string            timezone to use
  -V, --version                     display version and exit
```