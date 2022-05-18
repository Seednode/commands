/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	cockroach "seedno.de/seednode/commands-web/cockroach"
)

func ServePage() {

	var counter int = 1

	h1 := func(w http.ResponseWriter, _ *http.Request) {
		results, err := cockroach.RunQuery()
		if err != nil {
			panic(err)
		}
		io.WriteString(w, strconv.Itoa(counter)+"\n"+fmt.Sprintln(results))
		counter += 1
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
