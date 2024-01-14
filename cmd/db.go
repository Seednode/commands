/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	Url   string
	Table string
}

type Parameters struct {
	CommandCount int
	ExitCode     int
	HostName     string
	CommandName  string
	SortBy       string
	SortOrder    string
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

	host, err := GetEnvVar("COMMANDS_DB_HOST", DatabaseHost, false)
	if err != nil {
		return "", err
	}
	url.WriteString("host=" + host)

	port, err := GetEnvVar("COMMANDS_DB_PORT", DatabasePort, false)
	if err != nil {
		return "", err
	}
	url.WriteString(" port=" + port)

	user, err := GetEnvVar("COMMANDS_DB_USER", DatabaseUser, false)
	if err != nil {
		return "", err
	}
	url.WriteString(" user=" + user)

	if dbType == "postgresql" {
		pass, err := GetEnvVar("COMMANDS_DB_PASS", DatabasePass, true)
		if err != nil {
			return "", err
		}
		url.WriteString(" password=" + pass)
	}

	database, err := GetEnvVar("COMMANDS_DB_NAME", DatabaseName, false)
	if err != nil {
		return "", err
	}
	url.WriteString(" dbname=" + database)

	sslMode, err := GetEnvVar("COMMANDS_DB_SSL_MODE", DatabaseSslMode, false)
	if err != nil {
		return "", err
	}
	url.WriteString(" sslmode=" + sslMode)

	if dbType == "cockroachdb" {
		sslRootCert, err := GetEnvVar("COMMANDS_DB_ROOT_CERT", DatabaseRootCert, false)
		if err != nil {
			return "", err
		}
		url.WriteString(" sslrootcert=" + sslRootCert)

		sslClientKey, err := GetEnvVar("COMMANDS_DB_SSL_KEY", DatabaseSslKey, false)
		if err != nil {
			return "", err
		}
		url.WriteString(" sslkey=" + sslClientKey)

		sslClientCert, err := GetEnvVar("COMMANDS_DB_SSL_CERT", DatabaseSslCert, false)
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

func getRecentCommands(connection *pgx.Conn, tableName string, parameters *Parameters) ([]Row, error) {
	var rowSlice []Row

	var whereClauses = 0

	var statement strings.Builder

	statement.WriteString(fmt.Sprintf("%v\n%v\n%v\n%v\n%v\n%v\n%v\n%v %v",
		"select",
		"row_number() over() as row,",
		"date_trunc('second', starttime) as start_time,",
		"date_trunc('second', (age(stoptime, starttime)::time)) as duration,",
		"hostname as host_name,",
		"commandname as command_name,",
		"exitcode as exit_code",
		"from", tableName))

	if parameters.ExitCode != -1 {
		statement.WriteString(fmt.Sprintf("\nwhere exitcode = '%v'", parameters.ExitCode))
		whereClauses += 1
	}

	if parameters.HostName != "" {
		if whereClauses == 0 {
			statement.WriteString("\nwhere ")
		} else {
			statement.WriteString("\nand ")
		}
		statement.WriteString(fmt.Sprintf("hostname = '%v'", parameters.HostName))
		whereClauses += 1
	}

	if parameters.CommandName != "" {
		if whereClauses == 0 {
			statement.WriteString("\nwhere ")
		} else {
			statement.WriteString("\nand ")
		}
		statement.WriteString(fmt.Sprintf("commandname like '%%%v%%'", parameters.CommandName))
		whereClauses += 1
	}

	statement.WriteString(fmt.Sprintf("\norder by %v %v\n", parameters.SortBy, parameters.SortOrder))
	statement.WriteString(fmt.Sprintf("limit %v;", strconv.Itoa(parameters.CommandCount)))

	fmt.Printf("\n%s\n\n", statement.String())

	rows, err := connection.Query(context.Background(), statement.String())
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

func RunQuery(database *Database, parameters *Parameters) ([]Row, int, int, error) {
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

	commands, err := getRecentCommands(connection, database.Table, parameters)
	if err != nil {
		return []Row{}, 0, 0, err
	}

	return commands, totalCommandCount, failedCommandCount, nil
}
