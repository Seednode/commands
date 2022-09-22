/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cockroach

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
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

func GetDatabaseURL() (string, error) {
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

	return connection, nil
}

func closeDatabase(connection *pgx.Conn) error {
	err := connection.Close(context.Background())
	if err != nil {
		return err
	}

	return nil
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

	return totalCommandCount, nil
}

func getFailedCommandCount(connection *pgx.Conn) (int, error) {
	statement := "SELECT COUNT(exitcode) FROM logging WHERE exitcode <> 0"

	var failedCommandCount int
	err := connection.QueryRow(context.Background(), statement).Scan(&failedCommandCount)
	if err != nil {
		return failedCommandCount, err
	}

	return failedCommandCount, nil
}

func getRecentCommands(connection *pgx.Conn, commandCount int, exitCode int, hostName, commandName, sortBy, sortOrder string) ([]Row, error) {
	var rowSlice []Row

	var whereClauses = 0

	statement := fmt.Sprintf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v",
		"select",
		"row_number() over() as row,",
		"date_trunc('second', starttime) as start_time,",
		"date_trunc('second', (age(stoptime, starttime)::time)) as duration,",
		"hostname as host_name,",
		"commandname as command_name,",
		"exitcode as exit_code",
		"from logging")

	if exitCode != -1 {
		statement += fmt.Sprintf("\nwhere exitcode = '%v'", exitCode)
		whereClauses += 1
	}

	if hostName != "" {
		if whereClauses == 0 {
			statement += "\nwhere "
		} else {
			statement += "\nand "
		}
		statement += fmt.Sprintf("hostname = '%v'", hostName)
		whereClauses += 1
	}

	if commandName != "" {
		if whereClauses == 0 {
			statement += "\nwhere "
		} else {
			statement += "\nand "
		}
		statement += fmt.Sprintf("commandname like '%%%v%%'", commandName)
		whereClauses += 1
	}

	statement += fmt.Sprintf("\norder by %v %v\n", sortBy, sortOrder)
	statement += fmt.Sprintf("limit %v;", strconv.Itoa(commandCount))

	fmt.Printf("\n%v\n\n", statement)

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

	return rowSlice, nil
}

func RunQuery(databaseURL, timezone string, commandCount int, exitCode int, hostName, commandName, sortBy, sortOrder string) ([]Row, int, int, error) {

	connection, err := openDatabase(databaseURL)
	if err != nil {
		return []Row{}, 0, 0, err
	}
	defer func(connection *pgx.Conn) {
		err := closeDatabase(connection)
		if err != nil {
			fmt.Println(err)
		}
	}(connection)

	totalCommandCount, err := getTotalCommandCount(connection)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	failedCommandCount, err := getFailedCommandCount(connection)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	commands, err := getRecentCommands(connection, commandCount, exitCode, hostName, commandName, sortBy, sortOrder)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	for i := range commands {
		commands[i].StartTime, err = setTimeZone(commands[i].StartTime, timezone)
		if err != nil {
			return []Row{}, 0, 0, err
		}
	}

	return commands, totalCommandCount, failedCommandCount, nil
}
