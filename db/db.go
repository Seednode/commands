/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	utils "seedno.de/seednode/commands-web/utils"
)

type Database struct {
	Url   string
	Table string
}

type Row struct {
	RowNumber   int
	StartTime   time.Time
	Duration    time.Duration
	HostName    string
	CommandName string
	ExitCode    int
}

func GetDatabaseURL(dbType string) (string, error) {
	var url strings.Builder

	host, err := utils.GetEnvVar("COMMANDS_DB_HOST", false)
	if err != nil {
		return "", err
	}
	url.WriteString("host=" + host)

	port, err := utils.GetEnvVar("COMMANDS_DB_PORT", false)
	if err != nil {
		return "", err
	}
	url.WriteString(" port=" + port)

	user, err := utils.GetEnvVar("COMMANDS_DB_USER", false)
	if err != nil {
		return "", err
	}
	url.WriteString(" user=" + user)

	if dbType == "postgresql" {
		pass, err := utils.GetEnvVar("COMMANDS_DB_PASS", true)
		if err != nil {
			return "", err
		}
		url.WriteString(" password=" + pass)
	}

	database, err := utils.GetEnvVar("COMMANDS_DB_NAME", false)
	if err != nil {
		return "", err
	}
	url.WriteString(" dbname=" + database)

	sslMode, err := utils.GetEnvVar("COMMANDS_DB_SSL_MODE", false)
	if err != nil {
		return "", err
	}
	url.WriteString(" sslmode=" + sslMode)

	if dbType == "cockroachdb" {
		sslRootCert, err := utils.GetEnvVar("COMMANDS_DB_ROOT_CERT", false)
		if err != nil {
			return "", err
		}
		url.WriteString(" sslrootcert=" + sslRootCert)

		sslClientKey, err := utils.GetEnvVar("COMMANDS_DB_SSL_KEY", false)
		if err != nil {
			return "", err
		}
		url.WriteString(" sslkey=" + sslClientKey)

		sslClientCert, err := utils.GetEnvVar("COMMANDS_DB_SSL_CERT", false)
		if err != nil {
			return "", err
		}
		url.WriteString(" sslcert=" + sslClientCert)
	}

	return url.String(), nil
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

func getTotalCommandCount(connection *pgx.Conn, tableName string) (int, error) {
	statement := fmt.Sprintf("SELECT COUNT(commandname) FROM %s", tableName)

	var totalCommandCount int
	err := connection.QueryRow(context.Background(), statement).Scan(&totalCommandCount)
	if err != nil {
		return totalCommandCount, err
	}

	return totalCommandCount, nil
}

func getFailedCommandCount(connection *pgx.Conn, tableName string) (int, error) {
	statement := fmt.Sprintf("SELECT COUNT(exitcode) FROM %s WHERE exitcode <> 0", tableName)

	var failedCommandCount int
	err := connection.QueryRow(context.Background(), statement).Scan(&failedCommandCount)
	if err != nil {
		return failedCommandCount, err
	}

	return failedCommandCount, nil
}

func getRecentCommands(connection *pgx.Conn, tableName string, commandCount int, exitCode int, hostName, commandName, sortBy, sortOrder string) ([]Row, error) {
	var rowSlice []Row

	var whereClauses = 0

	statement := fmt.Sprintf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v %v",
		"select",
		"row_number() over() as row,",
		"date_trunc('second', starttime) as start_time,",
		"date_trunc('second', (age(stoptime, starttime)::time)) as duration,",
		"hostname as host_name,",
		"commandname as command_name,",
		"exitcode as exit_code",
		"from", tableName)

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

func RunQuery(database *Database, commandCount int, exitCode int, hostName, commandName, sortBy, sortOrder string) ([]Row, int, int, error) {
	connection, err := openDatabase(database.Url)
	if err != nil {
		return []Row{}, 0, 0, err
	}
	defer func(connection *pgx.Conn) {
		err := closeDatabase(connection)
		if err != nil {
			fmt.Println(err)
		}
	}(connection)

	totalCommandCount, err := getTotalCommandCount(connection, database.Table)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	failedCommandCount, err := getFailedCommandCount(connection, database.Table)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	commands, err := getRecentCommands(connection, database.Table, commandCount, exitCode, hostName, commandName, sortBy, sortOrder)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	for i := range commands {
		commands[i].StartTime = time.Now()
		if err != nil {
			return []Row{}, 0, 0, err
		}
	}

	return commands, totalCommandCount, failedCommandCount, nil
}
