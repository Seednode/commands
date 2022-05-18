/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cockroach

import (
	"context"
	"fmt"
	"time"

	pgx "github.com/jackc/pgx/v4"
	utils "seedno.de/seednode/commands-web/utils"
)

type Row struct {
	StartTime   time.Time
	Duration    time.Duration
	HostName    string
	CommandName string
	ExitCode    int
}

func getDatabaseURL() (string, error) {
	host, err := utils.GetEnvVar("ERRWRAPPER_DB_HOST")
	if err != nil {
		return "", err
	}
	host = "host=" + host

	port, err := utils.GetEnvVar("ERRWRAPPER_DB_PORT")
	if err != nil {
		return "", err
	}
	port = " port=" + port

	user, err := utils.GetEnvVar("ERRWRAPPER_DB_USER")
	if err != nil {
		return "", err
	}
	user = " user=" + user

	database, err := utils.GetEnvVar("ERRWRAPPER_DB_NAME")
	if err != nil {
		return "", err
	}
	database = " dbname=" + database

	sslMode, err := utils.GetEnvVar("ERRWRAPPER_DB_SSL_MODE")
	if err != nil {
		return "", err
	}
	sslMode = " sslmode=" + sslMode

	sslRootCert, err := utils.GetEnvVar("ERRWRAPPER_DB_ROOT_CERT")
	if err != nil {
		return "", err
	}
	sslRootCert = " sslrootcert=" + sslRootCert

	sslClientKey, err := utils.GetEnvVar("ERRWRAPPER_DB_SSL_KEY")
	if err != nil {
		return "", err
	}
	sslClientKey = " sslkey=" + sslClientKey

	sslClientCert, err := utils.GetEnvVar("ERRWRAPPER_DB_SSL_CERT")
	if err != nil {
		return "", err
	}
	sslClientCert = " sslcert=" + sslClientCert

	connection := fmt.Sprint(
		host,
		port,
		user,
		database,
		sslMode,
		sslRootCert,
		sslClientKey,
		sslClientCert,
	)

	return connection, nil
}

func openDatabase(databaseURL string) (*pgx.Conn, error) {
	connection, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func setTimeZone(oldTime time.Time) (time.Time, error) {
	timezone, err := utils.GetEnvVar("ERRWRAPPER_TZ")
	if err != nil {
		return oldTime, err
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return oldTime, err
	}

	newTime := oldTime.In(location)

	return newTime, nil
}

func createSQLStatement() string {
	statement := `select 
	date_trunc('second', starttime) as start_time,
	date_trunc('second', (age(stoptime, starttime)::time)) as duration,
	hostname as host_name,
	commandname as command_name,
	exitcode as exit_code
	from logging
	order by starttime desc
	limit 1000;`

	return statement
}

func getCommands(connection *pgx.Conn) ([]Row, error) {
	rowSlice := []Row{}

	statement := createSQLStatement()

	rows, err := connection.Query(context.Background(), statement)
	if err != nil {
		return rowSlice, err
	}
	defer rows.Close()

	for rows.Next() {
		var r Row
		err := rows.Scan(&r.StartTime, &r.Duration, &r.HostName, &r.CommandName, &r.ExitCode)
		if err != nil {
			return rowSlice, err
		}
		rowSlice = append(rowSlice, r)
	}

	return rowSlice, nil
}

func RunQuery() ([]Row, error) {
	err := utils.LoadEnv()
	if err != nil {
		return []Row{}, err
	}

	databaseURL, err := getDatabaseURL()
	if err != nil {
		return []Row{}, err
	}

	connection, err := openDatabase(databaseURL)
	if err != nil {
		return []Row{}, err
	}
	defer connection.Close(context.Background())

	commands, err := getCommands(connection)
	if err != nil {
		return []Row{}, err
	}

	for i := range commands {
		commands[i].StartTime, err = setTimeZone(commands[i].StartTime)
		if err != nil {
			return []Row{}, err
		}
	}

	return commands, nil
}
