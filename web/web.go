/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"

	cockroach "seedno.de/seednode/commands-web/cockroach"
)

func ServePage() {
	h1 := func(w http.ResponseWriter, _ *http.Request) {
		results, err := cockroach.RunQuery()
		if err != nil {
			panic(err)
		}
		io.WriteString(w, fmt.Sprintln(results))
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
