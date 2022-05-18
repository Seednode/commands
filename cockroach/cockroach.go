/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cockroach

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pgx "github.com/jackc/pgx/v4"
	utils "seedno.de/seednode/commands-web/utils"
)

type Row struct {
	RowNumber   int
	StartTime   time.Time
	Duration    time.Duration
	HostName    string
	CommandName string
	ExitCode    int
}

func getDatabaseURL() (string, error) {
	host, err := utils.GetEnvVar("COMMANDS_DB_HOST")
	if err != nil {
		return "", err
	}
	host = "host=" + host

	port, err := utils.GetEnvVar("COMMANDS_DB_PORT")
	if err != nil {
		return "", err
	}
	port = " port=" + port

	user, err := utils.GetEnvVar("COMMANDS_DB_USER")
	if err != nil {
		return "", err
	}
	user = " user=" + user

	database, err := utils.GetEnvVar("COMMANDS_DB_NAME")
	if err != nil {
		return "", err
	}
	database = " dbname=" + database

	sslMode, err := utils.GetEnvVar("COMMANDS_DB_SSL_MODE")
	if err != nil {
		return "", err
	}
	sslMode = " sslmode=" + sslMode

	sslRootCert, err := utils.GetEnvVar("COMMANDS_DB_ROOT_CERT")
	if err != nil {
		return "", err
	}
	sslRootCert = " sslrootcert=" + sslRootCert

	sslClientKey, err := utils.GetEnvVar("COMMANDS_DB_SSL_KEY")
	if err != nil {
		return "", err
	}
	sslClientKey = " sslkey=" + sslClientKey

	sslClientCert, err := utils.GetEnvVar("COMMANDS_DB_SSL_CERT")
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

	fmt.Printf("Set database URL to %v\n", connection)

	return connection, nil
}

func openDatabase(databaseURL string) (*pgx.Conn, error) {
	connection, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	fmt.Println("Opened database")

	return connection, nil
}

func setTimeZone(oldTime time.Time, timezone string) (time.Time, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return oldTime, err
	}

	newTime := oldTime.In(location)

	return newTime, nil
}

func getTotalCommandCount(connection *pgx.Conn) (int, error) {
	statement := "SELECT COUNT(commandname) FROM logging"

	var totalCommandCount int
	err := connection.QueryRow(context.Background(), statement).Scan(&totalCommandCount)
	if err != nil {
		return totalCommandCount, err
	}

	fmt.Printf("Retrieved total command count: %v\n", totalCommandCount)

	return totalCommandCount, nil
}

func getFailedCommandCount(connection *pgx.Conn) (int, error) {
	statement := "SELECT COUNT(exitcode) FROM logging WHERE exitcode <> 0"

	var failedCommandCount int
	err := connection.QueryRow(context.Background(), statement).Scan(&failedCommandCount)
	if err != nil {
		return failedCommandCount, err
	}

	fmt.Printf("Retrieved failed command count: %v\n", failedCommandCount)

	return failedCommandCount, nil
}

func getRecentCommands(connection *pgx.Conn, commandCount int) ([]Row, error) {
	rowSlice := []Row{}

	statement := `select
	row_number() over() as row,
	date_trunc('second', starttime) as start_time,
	date_trunc('second', (age(stoptime, starttime)::time)) as duration,
	hostname as host_name,
	commandname as command_name,
	exitcode as exit_code
	from logging
	order by starttime desc
	limit `
	statement += strconv.Itoa(commandCount) + ";"

	rows, err := connection.Query(context.Background(), statement)
	if err != nil {
		return rowSlice, err
	}
	defer rows.Close()

	for rows.Next() {
		var r Row
		err := rows.Scan(&r.RowNumber, &r.StartTime, &r.Duration, &r.HostName, &r.CommandName, &r.ExitCode)
		if err != nil {
			return rowSlice, err
		}
		rowSlice = append(rowSlice, r)
	}

	fmt.Printf("Retrieved %v recent commands.", len(rowSlice))

	return rowSlice, nil
}

func RunQuery(commandCount int) ([]Row, int, int, error) {
	err := utils.LoadEnv()
	if err != nil {
		fmt.Println("Environment file not found.")
	}

	databaseURL, err := getDatabaseURL()
	if err != nil {
		return []Row{}, 0, 0, err
	}

	connection, err := openDatabase(databaseURL)
	if err != nil {
		return []Row{}, 0, 0, err
	}
	defer connection.Close(context.Background())

	totalCommandCount, err := getTotalCommandCount(connection)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	failedCommandCount, err := getFailedCommandCount(connection)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	commands, err := getRecentCommands(connection, commandCount)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	timezone, err := utils.GetEnvVar("COMMANDS_TZ")
	if err != nil {
		return []Row{}, 0, 0, err
	}

	for i := range commands {
		commands[i].StartTime, err = setTimeZone(commands[i].StartTime, timezone)
		if err != nil {
			return []Row{}, 0, 0, err
		}
	}

	fmt.Println("Formatted time zones")

	return commands, totalCommandCount, failedCommandCount, nil
}
