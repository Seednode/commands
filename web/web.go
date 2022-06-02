/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"text/template"
	"time"

	cockroach "seedno.de/seednode/commands-web/cockroach"
	utils "seedno.de/seednode/commands-web/utils"
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

func ConstructPage(w io.Writer, databaseURL, timezone string, commandCount, exitCode int, hostName, commandName, sortBy, sortOrder string) error {
	startTime := time.Now()

	results, totalCommandCount, failedCommandCount, err := cockroach.RunQuery(databaseURL, timezone, commandCount, exitCode, hostName, commandName, sortBy, sortOrder)
	if err != nil {
		return err
	}

	t := template.New("t").Funcs(templateFuncs)
	t, err = t.Parse(htmlTemplate)
	if err != nil {
		return err
	}

	htmlHeader := GenerateHeader(commandCount, totalCommandCount, failedCommandCount)
	io.WriteString(w, htmlHeader)

	err = t.Execute(w, results)
	if err != nil {
		return err
	}

	htmlFooter := GenerateFooter()
	io.WriteString(w, htmlFooter)

	fmt.Printf("Constructed HTML page for up to %v commands (%v total, %v failed) in %v.\n",
		commandCount,
		totalCommandCount,
		failedCommandCount,
		time.Since(startTime))

	return nil
}

func servePageHandler(databaseURL, timezone string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		w.Header().Add("Content-Type", "text/html")
		err = ConstructPage(w, databaseURL, timezone, commandCount, exitCode, hostName, commandName, sortBy, sortOrder)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doNothing(w http.ResponseWriter, r *http.Request) {}

func ServePage() {
	err := utils.LoadEnv()
	if err != nil {
		fmt.Println("Environment file not found.")
		os.Exit(1)
	}

	databaseURL, err := cockroach.GetDatabaseURL()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	timezone, err := utils.GetEnvVar("COMMANDS_TZ")
	if err != nil {
		timezone = "UTC"
	}

	port, err := utils.GetEnvVar("COMMANDS_PORT")
	if err != nil {
		port = "8080"
	}

	http.HandleFunc("/", servePageHandler(databaseURL, timezone))
	http.HandleFunc("/favicon.ico", doNothing)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
