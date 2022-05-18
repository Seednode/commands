/*
Copyright © 2022 Seednode <seednode@seedno.de>
*/

package cmd

import (
	cockroach "seedno.de/seednode/commands-web/cockroach"
)

func ServePage() {

	_, err := cockroach.GetCommands()
	if err != nil {
		panic(err)
	}

	//	h1 := func(w http.ResponseWriter, _ *http.Request) {
	//		io.WriteString(w, starttime+"\n"+duration+"\n"+hostname+"\n"+commandname+"\n"+exitcode+"\n")
	//	}

	//	http.HandleFunc("/", h1)

	//	log.Fatal(http.ListenAndServe(":8080", nil))
}
