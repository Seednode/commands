/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"io"
	"log"
	"net/http"
	"reflect"
	"text/template"

	cockroach "seedno.de/seednode/commands-web/cockroach"
)

var templateFuncs = template.FuncMap{"rangeStruct": RangeStructer}

// In the template, we use rangeStruct to turn our struct values
// into a slice we can iterate over
var htmlTemplate = `{{range .}}<tr>
{{range rangeStruct .}} <td>{{.}}</td>
{{end}}</tr>
{{end}}`

// RangeStructer takes the first argument, which must be a struct, and
// returns the value of each field in a slice. It will return nil
// if there are no arguments or first argument is not a struct
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

func ConstructPage(w io.Writer) error {
	results, err := cockroach.RunQuery()
	if err != nil {
		return err
	}

	// We create the template and register out template function
	t := template.New("t").Funcs(templateFuncs)
	t, err = t.Parse(htmlTemplate)
	if err != nil {
		return err
	}

	htmlHeader := `
<html>
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
<table>
<thead><tr><th>start_time</th><th>duration</th><th>host_name</th><th>command_name</th><th>exit_code</th></tr></thead>
<tbody>
	`

	io.WriteString(w, htmlHeader)

	err = t.Execute(w, results)
	if err != nil {
		return err
	}

	htmlFooter := `
	</tbody>
	<tfoot><tr><td colspan=6>1000 rows</td></tr></tfoot></table>
	  </body>
	</html>
`

	io.WriteString(w, htmlFooter)

	return nil
}

func ServePage() {
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		ConstructPage(w)
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
