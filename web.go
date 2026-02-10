/*
Copyright © 2026 Seednode <seednode@seedno.de>
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"net/http/pprof"

	"github.com/julienschmidt/httprouter"
)

const (
	logDate string = `2006-01-02T15:04:05.000-07:00`
)

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

var htmlTemplate = `{{range .}}        <tr>
{{range rangeStruct .}}          <td>{{.}}</td>
{{end}}        </tr>
{{end}}`

func RangeStructer(args ...any) []any {
	if len(args) == 0 {
		return nil
	}

	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Struct {
		return nil
	}

	out := make([]any, v.NumField())
	for i := range v.NumField() {
		out[i] = v.Field(i).Interface()
	}

	return out
}

func securityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
	w.Header().Set("Permissions-Policy", "geolocation=(), midi=(), sync-xhr=(), microphone=(), camera=(), magnetometer=(), gyroscope=(), fullscreen=(), payment=()")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "1; mode=block")
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

	htmlHeader += fmt.Sprintf("  <h3>Displaying up to %v out of %v commands, including %v non-zero exit codes.</h3>",
		strconv.Itoa(commandCount),
		strconv.Itoa(totalCommandCount),
		strconv.Itoa(failedCommandCount))

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

func ConstructPage(w io.Writer, database *Database, parameters *Parameters) error {
	startTime := time.Now()

	results, totalCommandCount, failedCommandCount, err := RunQuery(database, parameters)
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

func ServePageHandler(database *Database) httprouter.Handle {
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

		parameters := &Parameters{
			CommandCount: commandCount,
			ExitCode:     exitCode,
			HostName:     hostName,
			CommandName:  commandName,
			SortBy:       sortBy,
			SortOrder:    sortOrder,
		}

		w.Header().Add("Content-Type", "text/html")

		securityHeaders(w)

		err = ConstructPage(w, database, parameters)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func ServerError(w http.ResponseWriter, r *http.Request, i any) {
	w.WriteHeader(http.StatusInternalServerError)

	w.Header().Add("Content-Type", "text/plain")

	securityHeaders(w)

	w.Write([]byte("500 Internal Server Error\n"))
}

func ServerErrorHandler() func(http.ResponseWriter, *http.Request, any) {
	return ServerError
}

func ServeVersion() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := fmt.Appendf(nil, "commands v%s\n", ReleaseVersion)

		securityHeaders(w)

		w.Header().Write(bytes.NewBufferString("Content-Length: " + strconv.Itoa(len(data))))

		w.Write(data)
	}
}

func ServePage() error {
	var err error

	timeZone := os.Getenv("TZ")
	if timeZone != "" {
		time.Local, err = time.LoadLocation(timeZone)
		if err != nil {
			return err
		}
	}

	if verbose {
		fmt.Printf("%s | START: commands v%s\n",
			time.Now().Format(logDate),
			ReleaseVersion,
		)
	}

	bindHost, err := net.LookupHost(bind)
	if err != nil {
		return err
	}

	bindAddr := net.ParseIP(bindHost[0])
	if bindAddr == nil {
		return errors.New("invalid bind address provided")
	}

	if databaseType != "cockroachdb" && databaseType != "postgresql" {
		return errors.New("invalid database type specified")
	}

	databaseURL, err := GetDatabaseURL()
	if err != nil {
		return err
	}

	database := &Database{
		Url:   databaseURL,
		Table: databaseTable,
	}

	mux := httprouter.New()

	mux.PanicHandler = ServerErrorHandler()

	mux.GET("/", ServePageHandler(database))

	mux.GET("/version", ServeVersion())

	if profile {
		mux.HandlerFunc("GET", "/debug/pprof/", pprof.Index)
		mux.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
		mux.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
		mux.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)
	}

	srv := &http.Server{
		Addr:         net.JoinHostPort(bind, strconv.Itoa(int(port))),
		Handler:      mux,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	if tlsKey != "" && tlsCert != "" {
		fmt.Printf("%s | Listening on %s://%s/\n",
			time.Now().Format(logDate),
			scheme,
			srv.Addr)

		err = srv.ListenAndServeTLS(tlsCert, tlsKey)
	} else {
		fmt.Printf("%s | Listening on %s://%s/\n",
			time.Now().Format(logDate),
			scheme,
			srv.Addr)

		err = srv.ListenAndServe()
	}

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
