/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package web

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"github.com/julienschmidt/httprouter"
	db "seedno.de/seednode/commands-web/db"
)

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

var htmlTemplate = `{{range .}}        <tr>
{{range rangeStruct .}}          <td>{{.}}</td>
{{end}}        </tr>
{{end}}`

func RangeStructer(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return nil
	}

	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Struct {
		return nil
	}

	out := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[i] = v.Field(i).Interface()
	}

	return out
}

func GenerateHeader(commandCount, totalCommandCount, failedCommandCount int) string {
	htmlHeader := `<html>
  <style>
    table {
      border: 2px solid #aaa;
      table-layout: fixed;
    }
    tr:nth-child(even) {
      background: #f4f4f4;
    }
    th,td {
      padding: 0.1em 0.5em;
    }
    td {
      border: 1px solid #aaa;
    }
    th {
      background: #eee;
      border: 1px solid #aaa;
      font-weight: bold;
      text-align: center;
    }
  </style>
  <head>
    <title>Command History</title>
  </head>
  <body>
  `

	htmlHeader += fmt.Sprintf("  <h3>Displaying up to %v out of %v commands, including %v non-zero exit codes.</h3>", strconv.Itoa(commandCount), strconv.Itoa(totalCommandCount), strconv.Itoa(failedCommandCount))

	htmlHeader += `
    <table>
      <thead>
        <tr>
          <th>row</th><th>start_time</th><th>duration</th><th>host_name</th><th>command_name</th><th>exit_code</th>
        </tr>
      </thead>
      <tbody>
`

	return htmlHeader
}

func GenerateFooter() string {
	htmlFooter := `      </tbody>
    </table>
  </body>
</html>`

	return htmlFooter
}

func ConstructPage(w io.Writer, database *db.Database, parameters *db.Parameters) error {
	startTime := time.Now()

	results, totalCommandCount, failedCommandCount, err := db.RunQuery(database, parameters)
	if err != nil {
		return err
	}

	t := template.New("t").Funcs(templateFuncs)
	t, err = t.Parse(htmlTemplate)
	if err != nil {
		return err
	}

	htmlHeader := GenerateHeader(parameters.CommandCount, totalCommandCount, failedCommandCount)
	_, err = io.WriteString(w, htmlHeader)
	if err != nil {
		return err
	}

	err = t.Execute(w, results)
	if err != nil {
		return err
	}

	htmlFooter := GenerateFooter()
	_, err = io.WriteString(w, htmlFooter)
	if err != nil {
		return err
	}

	fmt.Printf("Constructed HTML page for up to %v commands (%v total, %v failed) in %v.\n",
		parameters.CommandCount,
		totalCommandCount,
		failedCommandCount,
		time.Since(startTime))

	return nil
}

func ServePageHandler(database *db.Database) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var commandCount int
		commandCount, err := strconv.Atoi(r.URL.Query().Get("count"))
		if err != nil {
			commandCount = 1000
		}

		var exitCode int
		exitCode, err = strconv.Atoi(r.URL.Query().Get("exit_code"))
		if err != nil {
			exitCode = -1
		}

		hostName := r.URL.Query().Get("host_name")

		commandName := r.URL.Query().Get("command_name")

		sortBy := r.URL.Query().Get("sort_by")
		switch sortBy {
		case "duration":
			sortBy = "duration"
		case "host_name":
			sortBy = "hostname"
		case "command_name":
			sortBy = "commandname"
		case "exit_code":
			sortBy = "exitcode"
		default:
			sortBy = "starttime"
		}

		sortOrder := r.URL.Query().Get("sort_order")
		if sortOrder != "asc" {
			sortOrder = "desc"
		}

		parameters := &db.Parameters{
			CommandCount: commandCount,
			ExitCode:     exitCode,
			HostName:     hostName,
			CommandName:  commandName,
			SortBy:       sortBy,
			SortOrder:    sortOrder,
		}

		w.Header().Add("Content-Type", "text/html")
		err = ConstructPage(w, database, parameters)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func ServerError(w http.ResponseWriter, r *http.Request, i interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add("Content-Type", "text/plain")

	w.Write([]byte("500 Internal Server Error\n"))
}

func ServerErrorHandler() func(http.ResponseWriter, *http.Request, interface{}) {
	return ServerError
}
