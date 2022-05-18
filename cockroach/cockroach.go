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

func GetDatabaseURL() string {
	host := "host=" + utils.GetEnvVar("ERRWRAPPER_DB_HOST")
	port := " port=" + utils.GetEnvVar("ERRWRAPPER_DB_PORT")
	user := " user=" + utils.GetEnvVar("ERRWRAPPER_DB_USER")
	database := " dbname=" + utils.GetEnvVar("ERRWRAPPER_DB_NAME")
	sslMode := " sslmode=" + utils.GetEnvVar("ERRWRAPPER_DB_SSL_MODE")
	sslRootCert := " sslrootcert=" + utils.GetEnvVar("ERRWRAPPER_DB_ROOT_CERT")
	sslClientKey := " sslkey=" + utils.GetEnvVar("ERRWRAPPER_DB_SSL_KEY")
	sslClientCert := " sslcert=" + utils.GetEnvVar("ERRWRAPPER_DB_SSL_CERT")
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

	return connection
}

func CreateSQLStatement() (string, string) {
	statement1 := "set time zone 'America/Chicago';"
	statement2 := `select 
	date_trunc('second', starttime) as start_time,
	date_trunc('second', (age(stoptime, starttime)::time)) as duration,
	hostname as host_name,
	commandname as command_name,
	exitcode as exit_code
	from logging
	order by starttime desc
	limit 1000;`

	return statement1, statement2
}

func openDatabase(databaseURL string) (*pgx.Conn, error) {
	connection, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func GetCommands() ([]Row, error) {
	rowSlice := []Row{}

	err := utils.LoadEnv()
	if err != nil {
		return []Row{}, err
	}

	databaseURL := GetDatabaseURL()
	connection, err := openDatabase(databaseURL)
	if err != nil {
		return rowSlice, err
	}
	defer connection.Close(context.Background())

	statement1, statement2 := CreateSQLStatement()

	rows, err := connection.Query(context.Background(), statement1)
	if err != nil {
		return rowSlice, nil
	}
	rows.Close()

	rows, err = connection.Query(context.Background(), statement2)
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

	fmt.Println(rowSlice)

	return rowSlice, nil
}
