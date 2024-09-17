
## About
A basic frontend to view command history, with optional filtering via query parameters.

Feature requests, code criticism, bug reports, general chit-chat, and unrelated angst accepted at `commands@seedno.de`.

Static binary builds available [here](https://cdn.seedno.de/builds/commands).

x86_64 and ARM Docker images of latest version: `oci.seedno.de/seednode/commands:latest`.

Dockerfile available [here](https://github.com/Seednode/commands/blob/master/docker/Dockerfile).

Docker compose file example using Traefik available [here](https://github.com/Seednode/commands/blob/master/docker/docker-compose.yml).

<TODO: Add screenshot>

### Configuration
The following configuration methods are accepted, in order of highest to lowest priority:
- Command-line flags
- Environment variables

## Creating the table
This tool is designed for viewing the database generated by the [errwrapper](https://github.com/Seednode/errwrapper) tool, which should connect to the same database.

In this example, I'll be using the wonderful [usql](https://github.com/xo/usql) client.

### Connect to the database:
To connect, run the following, replacing the variables with their corresponding values:

`usql postgres://${COMMANDS_DB_USER}@${COMMANDS_DB_HOST}:${COMMANDS_DB_PORT}/${COMMANDS_DB_NAME}`

You should then be at a SQL prompt that looks something like the following:

`pg:commands@commands-db/logging=>`

### Create logging table
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

### Environment variables
Almost all options configurable via flags can also be configured via environment variables.

The associated environment variable is the prefix `COMMANDS_` plus the flag name, with the following changes:
- Leading hyphens removed
- Converted to upper-case
- All internal hyphens converted to underscores

For example:
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
  -b, --bind string           address to bind to (default "0.0.0.0")
      --db-host string        database host to connect to
      --db-name string        database name to connect to
      --db-pass string        database password to connect with
      --db-port string        database port to connect to
      --db-root-cert string   database ssl root certificate path
      --db-ssl-cert string    database ssl connection certificate path
      --db-ssl-key string     database ssl connection key path
      --db-ssl-mode string    database ssl connection mode
      --db-table string       database table to query
      --db-type string        database type to connect to
      --db-user string        database user to connect as
  -h, --help                  help for commands
  -p, --port uint16           port to listen on (default 8080)
      --profile               register net/http/pprof handlers
  -V, --version               display version and exit
```

## Building the Docker image
From inside the cloned repository, build the image using the following command:

`REGISTRY=<registry url> LATEST=yes TAG=alpine ./build-docker.sh`
